package models

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	dbpool *pgxpool.Pool
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
			return err
		}
	}

	dbpool, err = pgxpool.Connect(ctx, dbURL)
	if err != nil {
		return err
	}
	return nil
}

func Shutdown() {
	if dbpool != nil {
		dbpool.Close()
		dbpool = nil
	}
}
