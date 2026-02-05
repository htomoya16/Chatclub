package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"
)

type AnonymousChannelRepository interface {
	Upsert(ctx context.Context, ac domain.AnonymousChannel) error
	Delete(ctx context.Context, guildID, channelID string) error
	Get(ctx context.Context, guildID, channelID string) (*domain.AnonymousChannel, error)
}

type anonymousChannelRepository struct {
	db *sql.DB
}

func NewAnonymousChannelRepository(db *sql.DB) AnonymousChannelRepository {
	return &anonymousChannelRepository{db: db}
}

func (r *anonymousChannelRepository) Upsert(ctx context.Context, ac domain.AnonymousChannel) error {
	if ac.GuildID == "" || ac.ChannelID == "" {
		return errors.New("guildID and channelID are required")
	}

	// Ensure guild exists (minimal)
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO guilds (id) VALUES ($1)
         ON CONFLICT (id) DO NOTHING`,
		ac.GuildID,
	); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO anonymous_channels (guild_id, channel_id, webhook_id, webhook_token)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (guild_id, channel_id)
         DO UPDATE SET webhook_id = EXCLUDED.webhook_id,
                       webhook_token = EXCLUDED.webhook_token,
                       updated_at = now()`,
		ac.GuildID, ac.ChannelID, ac.WebhookID, ac.WebhookToken,
	)
	return err
}

func (r *anonymousChannelRepository) Delete(ctx context.Context, guildID, channelID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM anonymous_channels WHERE guild_id = $1 AND channel_id = $2`,
		guildID, channelID,
	)
	return err
}

func (r *anonymousChannelRepository) Get(ctx context.Context, guildID, channelID string) (*domain.AnonymousChannel, error) {
	var ac domain.AnonymousChannel
	err := r.db.QueryRowContext(ctx,
		`SELECT guild_id, channel_id, webhook_id, webhook_token, created_at, updated_at
         FROM anonymous_channels
         WHERE guild_id = $1 AND channel_id = $2`,
		guildID, channelID,
	).Scan(&ac.GuildID, &ac.ChannelID, &ac.WebhookID, &ac.WebhookToken, &ac.CreatedAt, &ac.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ac, nil
}
