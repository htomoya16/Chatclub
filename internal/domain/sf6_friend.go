package domain

import "time"

type SF6Friend struct {
	ID          string
	GuildID     string
	UserID      string
	FighterID   string
	DisplayName string
	Alias       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
