package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"ironpage-vault/internal/core"
)

// extractMentionUsernames keeps mention parsing centralized in core.
func extractMentionUsernames(text string) []string {
	return core.ExtractMentionUsernames(text)
}

// notifyMentionedUsersWithExecutor creates local notifications for annotation
// mentions inside the parent mutation's database boundary. A missing username is
// ignored as user input, while database and notification failures abort the
// mutation so a successful annotation cannot silently lose required side effects.
func (a *App) notifyMentionedUsersWithExecutor(ctx context.Context, executor sqlx.ExtContext, comment string, documentID string, authorID string) error {
	for _, username := range extractMentionUsernames(comment) {
		var userID string
		err := sqlx.GetContext(ctx, executor, &userID, `SELECT id FROM users WHERE username=$1`, piiLookupKey(a.cfg.AESKey, username))
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if userID == authorID {
			continue
		}
		if err := a.createNotificationWithExecutor(ctx, executor, userID, documentID, "annotation.mention", "You were mentioned in an annotation"); err != nil {
			return err
		}
	}
	return nil
}
