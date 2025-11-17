//go:build !exclude_migrations

package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00001, Down00001)
}

func Up00001(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
						CREATE TABLE events (
								id                UUID PRIMARY KEY,
								title             VARCHAR(128) NOT NULL,
								date_time         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
								event_duration    INTERVAL NOT NULL,
								description       VARCHAR(512),
								user_id           UUID         NOT NULL,
								notification_time TIMESTAMPTZ
						);
					`)
	return err
}

func Down00001(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS events;")
	return err
}
