package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00003, Down00003)
}

func Up00003(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		ALTER TYPE notification_status ADD VALUE 'PENDING_SENT';
	`)
	return err
}

func Down00003(_ context.Context, _ *sql.Tx) error {
	return nil
}
