package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"
)

type SF6FriendRepository interface {
	Upsert(ctx context.Context, friend domain.SF6Friend) error
	Delete(ctx context.Context, guildID, userID, fighterID string) error
	List(ctx context.Context, guildID, userID string) ([]domain.SF6Friend, error)
	ExistsByFighter(ctx context.Context, guildID, fighterID string) (bool, error)
}

type sf6FriendRepository struct {
	db *sql.DB
}

func NewSF6FriendRepository(db *sql.DB) SF6FriendRepository {
	return &sf6FriendRepository{db: db}
}

func (r *sf6FriendRepository) Upsert(ctx context.Context, friend domain.SF6Friend) error {
	if friend.GuildID == "" || friend.UserID == "" || friend.FighterID == "" {
		return errors.New("guildID, userID, fighterID are required")
	}
	if err := ensureGuildAndUser(ctx, r.db, friend.GuildID, friend.UserID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sf6_friends (guild_id, user_id, fighter_id, display_name, alias)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (guild_id, user_id, fighter_id)
         DO UPDATE SET display_name = EXCLUDED.display_name,
                       alias = EXCLUDED.alias,
                       updated_at = now()`,
		friend.GuildID, friend.UserID, friend.FighterID, friend.DisplayName, friend.Alias,
	)
	return err
}

func (r *sf6FriendRepository) Delete(ctx context.Context, guildID, userID, fighterID string) error {
	if guildID == "" || userID == "" || fighterID == "" {
		return errors.New("guildID, userID, fighterID are required")
	}
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sf6_friends WHERE guild_id = $1 AND user_id = $2 AND fighter_id = $3`,
		guildID, userID, fighterID,
	)
	return err
}

func (r *sf6FriendRepository) List(ctx context.Context, guildID, userID string) ([]domain.SF6Friend, error) {
	if guildID == "" || userID == "" {
		return nil, errors.New("guildID and userID are required")
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, guild_id, user_id, fighter_id, display_name, alias, created_at, updated_at
         FROM sf6_friends
         WHERE guild_id = $1 AND user_id = $2
         ORDER BY created_at ASC`,
		guildID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SF6Friend
	for rows.Next() {
		var friend domain.SF6Friend
		if err := rows.Scan(
			&friend.ID, &friend.GuildID, &friend.UserID, &friend.FighterID,
			&friend.DisplayName, &friend.Alias, &friend.CreatedAt, &friend.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, friend)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *sf6FriendRepository) ExistsByFighter(ctx context.Context, guildID, fighterID string) (bool, error) {
	if guildID == "" || fighterID == "" {
		return false, errors.New("guildID and fighterID are required")
	}
	row := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM sf6_friends WHERE guild_id = $1 AND fighter_id = $2 LIMIT 1`,
		guildID, fighterID,
	)
	var dummy int
	if err := row.Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
