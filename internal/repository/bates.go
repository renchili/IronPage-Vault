package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrBatesSequenceOverlap = errors.New("requested Bates start overlaps an allocated sequence")

// AllocateBatesRange locks the global sequence and reserves every page number
// used by one Bates job. The caller supplies its transaction so the reservation,
// job, document version, document pointer, and audit record commit together.
func AllocateBatesRange(ctx context.Context, executor sqlx.ExtContext, requested int, count int) (int, error) {
	if count < 1 {
		return 0, fmt.Errorf("Bates allocation count must be positive")
	}
	if _, err := executor.ExecContext(ctx, `INSERT INTO bates_sequences(scope,next_number,updated_at) VALUES('global',1,NOW()) ON CONFLICT(scope) DO NOTHING`); err != nil {
		return 0, err
	}
	var next int
	if err := sqlx.GetContext(ctx, executor, &next, `SELECT next_number FROM bates_sequences WHERE scope='global' FOR UPDATE`); err != nil {
		return 0, err
	}
	start := next
	if requested > 0 {
		if requested < next {
			return 0, ErrBatesSequenceOverlap
		}
		start = requested
	}
	if _, err := executor.ExecContext(ctx, `UPDATE bates_sequences SET next_number=$1,updated_at=NOW() WHERE scope='global'`, start+count); err != nil {
		return 0, err
	}
	return start, nil
}

func (r Repository) AllocateBatesStart(ctx context.Context, requested int) (int, error) {
	return AllocateBatesRange(ctx, r.DB, requested, 1)
}
