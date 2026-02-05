package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type SF6BattleRepository interface {
	Upsert(ctx context.Context, battle domain.SF6Battle) error
	BulkUpsert(ctx context.Context, battles []domain.SF6Battle) (int, error)
	ReassignOwnerBySubject(ctx context.Context, guildID, subjectFighterID, newUserID string) (int64, error)
	MarkOwnerKindUnlinkedBySubject(ctx context.Context, guildID, subjectFighterID string) (int64, error)
	ExistingSourceKeys(ctx context.Context, guildID, subjectFighterID string, keys []string) (map[string]struct{}, error)
	StatsByOpponentRange(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, startAt, endAt time.Time) ([]domain.SF6BattleStatRow, error)
	StatsByOpponentCount(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, limit int) ([]domain.SF6BattleStatRow, error)
	HistoryByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, limit, offset int) ([]domain.SF6BattleHistoryRow, error)
	CountByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string) (int, error)
	BattleTimesByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string) ([]time.Time, error)
	DeleteByUser(ctx context.Context, guildID, userID string) (int64, error)
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
	if battle.OwnerKind == "" {
		battle.OwnerKind = "account"
	}
	if err := ensureGuildAndUser(ctx, r.db, battle.GuildID, battle.UserID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sf6_battles (
            guild_id, user_id, owner_kind, subject_fighter_id, opponent_fighter_id, battle_at, result,
            self_character, opponent_character, round_wins, round_losses,
            source_key, session_id, raw_payload
         ) VALUES (
            $1, $2, $3, $4, $5, $6, $7,
            $8, $9, $10, $11,
            $12, $13, $14
         )
         ON CONFLICT (guild_id, subject_fighter_id, source_key)
         DO UPDATE SET user_id = EXCLUDED.user_id,
                       owner_kind = EXCLUDED.owner_kind,
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
		battle.OwnerKind,
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
            guild_id, user_id, owner_kind, subject_fighter_id, opponent_fighter_id, battle_at, result,
            self_character, opponent_character, round_wins, round_losses,
            source_key, session_id, raw_payload
         ) VALUES (
            $1, $2, $3, $4, $5, $6, $7,
            $8, $9, $10, $11,
            $12, $13, $14
         )
         ON CONFLICT (guild_id, subject_fighter_id, source_key)
         DO UPDATE SET user_id = EXCLUDED.user_id,
                       owner_kind = EXCLUDED.owner_kind,
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
		if battle.OwnerKind == "" {
			battle.OwnerKind = "account"
		}
		if _, err := stmt.ExecContext(ctx,
			battle.GuildID,
			battle.UserID,
			battle.OwnerKind,
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
         SET user_id = $1, owner_kind = 'account', updated_at = now()
         WHERE guild_id = $2 AND subject_fighter_id = $3`,
		newUserID, guildID, subjectFighterID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *sf6BattleRepository) MarkOwnerKindUnlinkedBySubject(ctx context.Context, guildID, subjectFighterID string) (int64, error) {
	if guildID == "" || subjectFighterID == "" {
		return 0, errors.New("guildID and subjectFighterID are required")
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE sf6_battles
         SET owner_kind = 'unlinked', updated_at = now()
         WHERE guild_id = $1 AND subject_fighter_id = $2`,
		guildID, subjectFighterID,
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

func (r *sf6BattleRepository) StatsByOpponentRange(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, startAt, endAt time.Time) ([]domain.SF6BattleStatRow, error) {
	if guildID == "" || subjectFighterID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, subjectFighterID, opponentFighterID are required")
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT self_character, result, COUNT(*)
         FROM sf6_battles
         WHERE guild_id = $1 AND subject_fighter_id = $2 AND opponent_fighter_id = $3
           AND battle_at >= $4 AND battle_at < $5
         GROUP BY self_character, result`,
		guildID, subjectFighterID, opponentFighterID, startAt, endAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.SF6BattleStatRow
	for rows.Next() {
		var row domain.SF6BattleStatRow
		if err := rows.Scan(&row.SelfCharacter, &row.Result, &row.Count); err != nil {
			return nil, err
		}
		stats = append(stats, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *sf6BattleRepository) StatsByOpponentCount(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, limit int) ([]domain.SF6BattleStatRow, error) {
	if guildID == "" || subjectFighterID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, subjectFighterID, opponentFighterID are required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT self_character, result, COUNT(*)
         FROM (
            SELECT self_character, result
            FROM sf6_battles
            WHERE guild_id = $1 AND subject_fighter_id = $2 AND opponent_fighter_id = $3
            ORDER BY battle_at DESC
            LIMIT $4
         ) AS recent
         GROUP BY self_character, result`,
		guildID, subjectFighterID, opponentFighterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.SF6BattleStatRow
	for rows.Next() {
		var row domain.SF6BattleStatRow
		if err := rows.Scan(&row.SelfCharacter, &row.Result, &row.Count); err != nil {
			return nil, err
		}
		stats = append(stats, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *sf6BattleRepository) HistoryByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string, limit, offset int) ([]domain.SF6BattleHistoryRow, error) {
	if guildID == "" || subjectFighterID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, subjectFighterID, opponentFighterID are required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	if offset < 0 {
		return nil, errors.New("offset must be >= 0")
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT battle_at, result, self_character, opponent_character
         FROM sf6_battles
         WHERE guild_id = $1 AND subject_fighter_id = $2 AND opponent_fighter_id = $3
         ORDER BY battle_at DESC, source_key DESC
         LIMIT $4 OFFSET $5`,
		guildID, subjectFighterID, opponentFighterID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SF6BattleHistoryRow
	for rows.Next() {
		var row domain.SF6BattleHistoryRow
		if err := rows.Scan(&row.BattleAt, &row.Result, &row.SelfCharacter, &row.OpponentCharacter); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *sf6BattleRepository) CountByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string) (int, error) {
	if guildID == "" || subjectFighterID == "" || opponentFighterID == "" {
		return 0, errors.New("guildID, subjectFighterID, opponentFighterID are required")
	}
	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*)
         FROM sf6_battles
         WHERE guild_id = $1 AND subject_fighter_id = $2 AND opponent_fighter_id = $3`,
		guildID, subjectFighterID, opponentFighterID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *sf6BattleRepository) BattleTimesByOpponent(ctx context.Context, guildID, subjectFighterID, opponentFighterID string) ([]time.Time, error) {
	if guildID == "" || subjectFighterID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, subjectFighterID, opponentFighterID are required")
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT battle_at
         FROM sf6_battles
         WHERE guild_id = $1 AND subject_fighter_id = $2 AND opponent_fighter_id = $3
         ORDER BY battle_at DESC, source_key DESC`,
		guildID, subjectFighterID, opponentFighterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *sf6BattleRepository) DeleteByUser(ctx context.Context, guildID, userID string) (int64, error) {
	if guildID == "" || userID == "" {
		return 0, errors.New("guildID and userID are required")
	}
	if err := ensureGuildAndUser(ctx, r.db, guildID, userID); err != nil {
		return 0, err
	}
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM sf6_battles WHERE guild_id = $1 AND user_id = $2`,
		guildID, userID,
	)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
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
