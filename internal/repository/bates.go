package repository

import "context"

func (r Repository) AllocateBatesStart(ctx context.Context, requested int) (int, error) {
	if requested > 0 {
		return requested, nil
	}
	var start int
	if err := r.DB.GetContext(ctx, &start, `INSERT INTO bates_sequences(scope,next_number,updated_at) VALUES('global',2,NOW()) ON CONFLICT(scope) DO UPDATE SET next_number=bates_sequences.next_number+1, updated_at=NOW() RETURNING next_number-1`); err != nil {
		return 0, err
	}
	return start, nil
}
