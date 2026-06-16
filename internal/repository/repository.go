package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return Repository{DB: db}
}

func (r Repository) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}
