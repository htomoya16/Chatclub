package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"
)

type SF6SessionRepository interface {
	Start(ctx context.Context, guildID, userID, opponentFighterID string, startedAt time.Time) (*domain.SF6Session, error)
	GetActive(ctx context.Context, guildID, userID, opponentFighterID string) (*domain.SF6Session, error)
	End(ctx context.Context, guildID, userID, opponentFighterID string, endedAt time.Time) (*domain.SF6Session, error)
}

type sf6SessionRepository struct {
	db *sql.DB
}

func NewSF6SessionRepository(db *sql.DB) SF6SessionRepository {
	return &sf6SessionRepository{db: db}
}

func (r *sf6SessionRepository) Start(ctx context.Context, guildID, userID, opponentFighterID string, startedAt time.Time) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	if err := ensureGuildAndUser(ctx, r.db, guildID, userID); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE sf6_sessions
         SET status = 'ended', ended_at = $3, updated_at = now()
         WHERE guild_id = $1 AND user_id = $2 AND status = 'active'`,
		guildID, userID, startedAt,
	); err != nil {
		return nil, err
	}

	var session domain.SF6Session
	row := tx.QueryRowContext(ctx,
		`INSERT INTO sf6_sessions (
            guild_id, user_id, opponent_fighter_id, status, started_at, last_polled_at
         ) VALUES (
            $1, $2, $3, 'active', $4, $4
         )
         RETURNING id, guild_id, user_id, opponent_fighter_id, status, started_at`,
		guildID, userID, opponentFighterID, startedAt,
	)
	if err := row.Scan(&session.ID, &session.GuildID, &session.UserID, &session.OpponentFighterID, &session.Status, &session.StartedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sf6SessionRepository) GetActive(ctx context.Context, guildID, userID, opponentFighterID string) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	row := r.db.QueryRowContext(ctx,
		`SELECT id, guild_id, user_id, opponent_fighter_id, status, started_at, ended_at
         FROM sf6_sessions
         WHERE guild_id = $1 AND user_id = $2 AND opponent_fighter_id = $3 AND status = 'active'
         ORDER BY started_at DESC
         LIMIT 1`,
		guildID, userID, opponentFighterID,
	)
	var session domain.SF6Session
	if err := row.Scan(&session.ID, &session.GuildID, &session.UserID, &session.OpponentFighterID, &session.Status, &session.StartedAt, &session.EndedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sf6SessionRepository) End(ctx context.Context, guildID, userID, opponentFighterID string, endedAt time.Time) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	session, err := r.GetActive(ctx, guildID, userID, opponentFighterID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	row := r.db.QueryRowContext(ctx,
		`UPDATE sf6_sessions
         SET status = 'ended', ended_at = $2, updated_at = now()
         WHERE id = $1
         RETURNING id, guild_id, user_id, opponent_fighter_id, status, started_at, ended_at`,
		session.ID, endedAt,
	)
	var updated domain.SF6Session
	if err := row.Scan(&updated.ID, &updated.GuildID, &updated.UserID, &updated.OpponentFighterID, &updated.Status, &updated.StartedAt, &updated.EndedAt); err != nil {
		return nil, err
	}
	return &updated, nil
}
