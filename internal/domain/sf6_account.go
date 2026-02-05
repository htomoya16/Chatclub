package domain

import "time"

type SF6Account struct {
	ID          string
	GuildID     string
	UserID      string
	FighterID   string
	DisplayName string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
