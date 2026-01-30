package domain

import (
	"encoding/json"
	"time"
)

type SF6Battle struct {
	ID                string
	GuildID           string
	UserID            string
	OwnerKind         string
	SubjectFighterID  string
	OpponentFighterID string
	BattleAt          time.Time
	Result            string
	SelfCharacter     string
	OpponentCharacter string
	RoundWins         int
	RoundLosses       int
	SourceKey         string
	SessionID         *string
	RawPayload        json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
