package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00002, Down00002)
}

func Up00002(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
						CREATE TYPE notification_status AS ENUM ('PENDING', 'SENT');

						ALTER TABLE events
						ADD COLUMN notification_status notification_status;
					`)
	return err
}

func Down00002(_ context.Context, _ *sql.Tx) error {
	return nil
}
