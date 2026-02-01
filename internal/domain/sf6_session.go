package domain

import "time"

type SF6Session struct {
	ID                string
	GuildID           string
	UserID            string
	OpponentFighterID string
	Status            string
	StartedAt         time.Time
	EndedAt           *time.Time
}
