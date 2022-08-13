package models

import (
	"context"
	"database/sql"
	"time"
)

// Session describes an active user session
type Session struct {
	ID       string
	User     User
	Company  Company
	LastSeen time.Time
}

// SessionRemove removes a user session from the list of active sessions
func SessionRemove(ctx context.Context, sessionID string) error {
	tx := getContextTx(ctx)
	query := `DELETE FROM session WHERE id = $1`

	_, err := tx.Exec(ctx, query, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// SessionGet checks if the supplied sessionID is active
func SessionGet(ctx context.Context, sessionID string) (Session, error) {

	tx := getContextTx(ctx)
	var s Session

	query := `SELECT id, user_id AS "user.id",
COALESCE(company_id, 0) AS "company.id",
last_seen
FROM session WHERE id = $1`
	err := tx.Get(ctx, &s, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return s, nil
		}
		return s, err
	}

	s.User, err = UserGet(ctx, UserFilter{ID: s.User.ID})
	if err != nil {
		return s, err
	}

	if s.Company.ID > 0 {
		s.Company, err = CompanyGet(ctx, CompanyFilter{ID: s.Company.ID})
		if err != nil {
			return s, err
		}
	}

	return s, nil
}

// SessionSave creates or updates a session in the database
func SessionSave(ctx context.Context, s Session) (string, error) {
	var query string
	s.LastSeen = time.Now()
	tx := getContextTx(ctx)

	if s.ID == "" {
		id, err := GenerateRandomString(20)
		if err != nil {
			return "", err
		}

		s.ID = string(id)

		query = `INSERT INTO session (id, user_id, company_id, last_seen) VALUES (:id, :user.id, :company.id, :last_seen)`
		if s.Company.ID == 0 {
			query = `INSERT INTO session (id, user_id, last_seen) VALUES (:id, :user.id, :last_seen)`
		}
	} else {
		query = `UPDATE session SET company_id = :company.id, last_seen = :last_seen WHERE id = :id`
	}

	_, err := tx.NamedExec(ctx, query, s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}
