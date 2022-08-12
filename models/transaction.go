package models

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/yzzyx/faktura-pdf/sqlx"
	"github.com/yzzyx/zerr"
)

// Usage:
// Transactions are stored in the context supplied to all Datastore functions.
//
// Usage in controllers:
// =====================
// In order to start a new transaction, add the following lines to the calling function:
//
// ---
// ctx, err := dstx.Begin(request.Context())
// if err != nil {
//     // handle error
// }
// defer func() {
//   dstx.CommitOrRollback(ctx, err)
// }()
// ---
//
// The 'Begin()' method starts a new transaction, that will be used by datastore functions.
// In order to handle Commit/Rollback on error, defer a function that calls 'CommitOrRollback()',
// with both context and error supplied. If error is set, the whole transaction will be rolled back,
// and all messages will be discarded. If no error is set, the transaction will be commited, and all
// queued messages will be published to the messagequeue.
//
// Usage in datastores
// ===================
// When using transactions in datastores, the transaction must be read from the supplied context.
// This is performed by calling `getContextTx(ctx, db)`. If a transaction is currently in progress,
// it will be returned. Otherwise the supplied database handle will be returned.
//
// Example
// func (d *DatastoreExample) List(ctx context.Context, filter models.ExampleFilter) ([]models.Example, error) {
//   tx := getContextTx(ctx, d.DB)
//   tx.NamedQuery("SELECT 1", filter)
//   ...
//   ...
// }

// transaction describes a transaction and implements a queue for messages
type transaction struct {
	tx *sqlx.PgxTx
}

// Begin starts a new transaction
func Begin(ctx context.Context) (context.Context, error) {
	var err error

	tx := &transaction{}
	c := sqlx.NewPgxPool(dbpool)
	tx.tx, err = c.Beginx(ctx)
	if err != nil {
		return nil, zerr.Wrap(err)
	}

	ctx = context.WithValue(ctx, TransactionContextKey, tx)
	return ctx, nil
}

// CommitOrRollback checks if error is set, and commits or rollbacks a transaction and publishes all messages in queue
// This _must_ be called in a defer as:
//   defer tx.CommitOrRollback(ctx, &err)
// Otherwise we can't both catch panics and errors.
// See here for an example: https://gist.github.com/yzzyx/584450948c5cf25eaad00013dfb43b7d
func CommitOrRollback(ctx context.Context, errPtr *error) {
	var err error
	if errPtr != nil {
		err = *errPtr
	}

	txVal := ctx.Value(TransactionContextKey)
	if txVal == nil {
		//if logger != nil {
		//	zerr.Wrap(errors.New("CommitOrRollBack called without ongoing transaction")).LogError(d.Logger)
		//}
		log.Printf("CommitOrRollback: %+v", zerr.Wrap(errors.New("CommitOrRollBack called without ongoing transaction")))
		return
	}

	tx, ok := txVal.(*transaction)
	if !ok {
		//if logger != nil {
		//	zerr.Wrap(errors.New("TransactionContextKey does not contain transaction")).LogError(d.Logger)
		//}
		log.Printf("CommitOrRollback: %+v", zerr.Wrap(errors.New("TransactionContextKey does not contain transaction")))
		return
	}

	if r := recover(); r != nil {
		// A panic occurred - set error, which will rollback transaction
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}

		//if logger != nil {
		//	zerr.Wrap(err).LogError(d.Logger)
		//} else {
		//	fmt.Fprintln(os.Stderr, "panic():", err)
		//	debug.PrintStack()
		//}
		log.Printf("CommitOrRollback: %+v", zerr.Wrap(err))
	}

	if err != nil {
		// Rollback transaction and discard messages
		txerr := tx.tx.Rollback(ctx)
		if txerr != nil {
			//if d.Logger != nil {
			//	zerr.Wrap(txerr).LogError(d.Logger)
			//}
			log.Printf("CommitOrRollback: %+v", zerr.Wrap(txerr))
		}
		return
	}

	// If no error was passed in, commit transaction and publish messages
	txerr := tx.tx.Commit(ctx)
	if txerr != nil {
		//if d.Logger != nil {
		//	zerr.Wrap(txerr).LogError(d.Logger)
		//}
		log.Printf("CommitOrRollback: %+v", zerr.Wrap(txerr))
		return
	}
}

// getContextTx returns the currently active transaction
func getContextTx(ctx context.Context) *sqlx.PgxTx {
	txVal := ctx.Value(TransactionContextKey)
	tx, ok := txVal.(*transaction)
	if txVal == nil || !ok {
		panic("no active transaction - cannot continue")
	}
	return tx.tx
}

// GetTx returns any active transaction or the supplied database ptr
func GetTx(ctx context.Context) *sqlx.PgxTx {
	return getContextTx(ctx)
}

// statusResponseWriter implements the interface corresponding to http.ResponseWriter,
// but also saves the statusCode set on the responsewriter
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader sets the response status code, and saves it for reading later
func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// TransactionMiddleware creates a new transaction that is used by all wrapped controllers
func TransactionMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var err error
		ctx := r.Context()

		// Begin a new transaction
		ctx, err = Begin(r.Context())
		if err != nil {
			log.Printf("TransactionMiddleware - error occurred when starting transaction: %+v", err)
			return
		}

		defer CommitOrRollback(ctx, &err)

		r = r.WithContext(ctx)
		srw := &statusResponseWriter{w, http.StatusOK}
		next.ServeHTTP(srw, r)

		// All statuscodes above 400 implies either client errors or server errors.
		// This means that we should set 'err' appropriately, which will in turn rollback transaction
		if srw.statusCode > 400 {
			err = errors.New("error statuscode")
		}
	}
	return http.HandlerFunc(fn)
}
