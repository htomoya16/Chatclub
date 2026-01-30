package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type SF6BattleRepository interface {
	Upsert(ctx context.Context, battle domain.SF6Battle) error
	BulkUpsert(ctx context.Context, battles []domain.SF6Battle) (int, error)
	ReassignOwnerBySubject(ctx context.Context, guildID, subjectFighterID, newUserID string) (int64, error)
	ExistingSourceKeys(ctx context.Context, guildID, subjectFighterID string, keys []string) (map[string]struct{}, error)
}

type sf6BattleRepository struct {
	db *sql.DB
}

func NewSF6BattleRepository(db *sql.DB) SF6BattleRepository {
	return &sf6BattleRepository{db: db}
}

func (r *sf6BattleRepository) Upsert(ctx context.Context, battle domain.SF6Battle) error {
	if battle.GuildID == "" || battle.UserID == "" || battle.SubjectFighterID == "" || battle.OpponentFighterID == "" {
		return errors.New("guildID, userID, subjectFighterID, opponentFighterID are required")
	}
	if battle.SourceKey == "" {
		return errors.New("sourceKey is required")
	}
	if err := ensureGuildAndUser(ctx, r.db, battle.GuildID, battle.UserID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sf6_battles (
            guild_id, user_id, subject_fighter_id, opponent_fighter_id, battle_at, result,
            self_character, opponent_character, round_wins, round_losses,
            source_key, session_id, raw_payload
         ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10,
            $11, $12, $13
         )
         ON CONFLICT (guild_id, subject_fighter_id, source_key)
         DO UPDATE SET user_id = EXCLUDED.user_id,
                       opponent_fighter_id = EXCLUDED.opponent_fighter_id,
                       battle_at = EXCLUDED.battle_at,
                       result = EXCLUDED.result,
                       self_character = EXCLUDED.self_character,
                       opponent_character = EXCLUDED.opponent_character,
                       round_wins = EXCLUDED.round_wins,
                       round_losses = EXCLUDED.round_losses,
                       session_id = EXCLUDED.session_id,
                       raw_payload = EXCLUDED.raw_payload,
                       updated_at = now()`,
		battle.GuildID,
		battle.UserID,
		battle.SubjectFighterID,
		battle.OpponentFighterID,
		battle.BattleAt,
		battle.Result,
		battle.SelfCharacter,
		battle.OpponentCharacter,
		battle.RoundWins,
		battle.RoundLosses,
		battle.SourceKey,
		battle.SessionID,
		nullIfEmptyBytes(battle.RawPayload),
	)
	return err
}

func (r *sf6BattleRepository) BulkUpsert(ctx context.Context, battles []domain.SF6Battle) (int, error) {
	if len(battles) == 0 {
		return 0, nil
	}
	first := battles[0]
	if err := ensureGuildAndUser(ctx, r.db, first.GuildID, first.UserID); err != nil {
		return 0, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO sf6_battles (
            guild_id, user_id, subject_fighter_id, opponent_fighter_id, battle_at, result,
            self_character, opponent_character, round_wins, round_losses,
            source_key, session_id, raw_payload
         ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10,
            $11, $12, $13
         )
         ON CONFLICT (guild_id, subject_fighter_id, source_key)
         DO UPDATE SET user_id = EXCLUDED.user_id,
                       opponent_fighter_id = EXCLUDED.opponent_fighter_id,
                       battle_at = EXCLUDED.battle_at,
                       result = EXCLUDED.result,
                       self_character = EXCLUDED.self_character,
                       opponent_character = EXCLUDED.opponent_character,
                       round_wins = EXCLUDED.round_wins,
                       round_losses = EXCLUDED.round_losses,
                       session_id = EXCLUDED.session_id,
                       raw_payload = EXCLUDED.raw_payload,
                       updated_at = now()`,
	)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, battle := range battles {
		if battle.GuildID == "" || battle.UserID == "" || battle.SubjectFighterID == "" || battle.OpponentFighterID == "" || battle.SourceKey == "" {
			return 0, errors.New("battle has missing required fields")
		}
		if _, err := stmt.ExecContext(ctx,
			battle.GuildID,
			battle.UserID,
			battle.SubjectFighterID,
			battle.OpponentFighterID,
			battle.BattleAt,
			battle.Result,
			battle.SelfCharacter,
			battle.OpponentCharacter,
			battle.RoundWins,
			battle.RoundLosses,
			battle.SourceKey,
			battle.SessionID,
			nullIfEmptyBytes(battle.RawPayload),
		); err != nil {
			return 0, err
		}
		count++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *sf6BattleRepository) ReassignOwnerBySubject(ctx context.Context, guildID, subjectFighterID, newUserID string) (int64, error) {
	if guildID == "" || subjectFighterID == "" || newUserID == "" {
		return 0, errors.New("guildID, subjectFighterID, newUserID are required")
	}
	if err := ensureGuildAndUser(ctx, r.db, guildID, newUserID); err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE sf6_battles
         SET user_id = $1, updated_at = now()
         WHERE guild_id = $2 AND subject_fighter_id = $3`,
		newUserID, guildID, subjectFighterID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *sf6BattleRepository) ExistingSourceKeys(ctx context.Context, guildID, subjectFighterID string, keys []string) (map[string]struct{}, error) {
	if guildID == "" || subjectFighterID == "" {
		return nil, errors.New("guildID and subjectFighterID are required")
	}
	unique := uniqueStrings(keys)
	if len(unique) == 0 {
		return map[string]struct{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT source_key
         FROM sf6_battles
         WHERE guild_id = $1 AND subject_fighter_id = $2 AND source_key = ANY($3)`,
		guildID, subjectFighterID, pq.Array(unique),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exists := make(map[string]struct{}, len(unique))
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		exists[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return exists, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func nullIfEmpty(value string) sql.NullString {
	if value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullIfEmptyBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return value
}
