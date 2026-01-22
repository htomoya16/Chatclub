package repository

import (
	"context"
	"database/sql"
	"time"
)

type HealthRepository interface {
	PingDB(ctx context.Context) error
}

type healthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) HealthRepository {
	return &healthRepository{db: db}
}

// DBへ到達できるか
func (r *healthRepository) PingDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	return r.db.PingContext(ctx)
}
