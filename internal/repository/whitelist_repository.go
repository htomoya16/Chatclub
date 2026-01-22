package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
)

type WhitelistRepository interface {
	Add(ctx context.Context, w domain.Whitelist) error
	Remove(ctx context.Context, platform, userID string) error
	Exists(ctx context.Context, platform, userID string) (bool, error)
	List(ctx context.Context, platform string, limit, offset int) ([]domain.Whitelist, error)
}

type whitelistRepository struct {
	db *sql.DB
}

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

func (r *whitelistRepository) Add(ctx context.Context, w domain.Whitelist) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO whitelists (platform, user_id, note)
         VALUES (?, ?, ?)
         ON DUPLICATE KEY UPDATE note = VALUES(note)`,
		w.Platform, w.UserID, w.Note,
	)
	return err
}

func (r *whitelistRepository) Remove(ctx context.Context, platform, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM whitelists WHERE platform = ? AND user_id = ?`,
		platform, userID,
	)
	return err
}

func (r *whitelistRepository) Exists(ctx context.Context, platform, userID string) (bool, error) {
	var dummy int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM whitelists WHERE platform = ? AND user_id = ? LIMIT 1`,
		platform, userID,
	).Scan(&dummy)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *whitelistRepository) List(ctx context.Context, platform string, limit, offset int) ([]domain.Whitelist, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, platform, user_id, note, created_at
		   FROM whitelists
		  WHERE platform = ?
		  ORDER BY id DESC
		  LIMIT ? OFFSET ?`,
		platform, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Whitelist
	for rows.Next() {
		var w domain.Whitelist
		if err := rows.Scan(&w.ID, &w.Platform, &w.UserID, &w.Note, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
