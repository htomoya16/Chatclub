package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"
)

type SF6AccountRepository interface {
	Upsert(ctx context.Context, account domain.SF6Account) error
	GetByUser(ctx context.Context, guildID, userID string) (*domain.SF6Account, error)
	GetByFighter(ctx context.Context, guildID, fighterID string) (*domain.SF6Account, error)
}

type sf6AccountRepository struct {
	db *sql.DB
}

func NewSF6AccountRepository(db *sql.DB) SF6AccountRepository {
	return &sf6AccountRepository{db: db}
}

func (r *sf6AccountRepository) Upsert(ctx context.Context, account domain.SF6Account) error {
	if account.GuildID == "" || account.UserID == "" || account.FighterID == "" {
		return errors.New("guildID, userID, fighterID are required")
	}
	if err := ensureGuildAndUser(ctx, r.db, account.GuildID, account.UserID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sf6_accounts (guild_id, user_id, fighter_id, display_name, status)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (guild_id, user_id)
         DO UPDATE SET fighter_id = EXCLUDED.fighter_id,
                       display_name = EXCLUDED.display_name,
                       status = EXCLUDED.status,
                       updated_at = now()`,
		account.GuildID, account.UserID, account.FighterID, account.DisplayName, account.Status,
	)
	return err
}

func (r *sf6AccountRepository) GetByUser(ctx context.Context, guildID, userID string) (*domain.SF6Account, error) {
	if guildID == "" || userID == "" {
		return nil, errors.New("guildID and userID are required")
	}
	var account domain.SF6Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, guild_id, user_id, fighter_id, display_name, status, created_at, updated_at
         FROM sf6_accounts
         WHERE guild_id = $1 AND user_id = $2`,
		guildID, userID,
	).Scan(
		&account.ID, &account.GuildID, &account.UserID, &account.FighterID,
		&account.DisplayName, &account.Status, &account.CreatedAt, &account.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *sf6AccountRepository) GetByFighter(ctx context.Context, guildID, fighterID string) (*domain.SF6Account, error) {
	if guildID == "" || fighterID == "" {
		return nil, errors.New("guildID and fighterID are required")
	}
	var account domain.SF6Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, guild_id, user_id, fighter_id, display_name, status, created_at, updated_at
         FROM sf6_accounts
         WHERE guild_id = $1 AND fighter_id = $2`,
		guildID, fighterID,
	).Scan(
		&account.ID, &account.GuildID, &account.UserID, &account.FighterID,
		&account.DisplayName, &account.Status, &account.CreatedAt, &account.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}
