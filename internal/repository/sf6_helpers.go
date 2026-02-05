package repository

import (
	"context"
	"database/sql"
	"errors"
)

func ensureGuildAndUser(ctx context.Context, db *sql.DB, guildID, userID string) error {
	if guildID == "" || userID == "" {
		return errors.New("guildID and userID are required")
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO guilds (id) VALUES ($1)
         ON CONFLICT (id) DO NOTHING`,
		guildID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO users (id) VALUES ($1)
         ON CONFLICT (id) DO NOTHING`,
		userID,
	); err != nil {
		return err
	}
	return nil
}
