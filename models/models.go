package models

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yzzyx/zerr"
)

var (
	dbpool *pgxpool.Pool
)

var (
	errTooManyRows = errors.New("too many rows returned")
)

func Setup(ctx context.Context, dbURL string) error {
	var err error
	m, err := migrate.New(
		"file://migrations",
		dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return zerr.Wrap(err)
		}
	}

	dbpool, err = pgxpool.Connect(ctx, dbURL)
	if err != nil {
		return zerr.Wrap(err).WithString("url", dbURL)
	}
	return nil
}

func Shutdown() {
	if dbpool != nil {
		dbpool.Close()
		dbpool = nil
	}
}
