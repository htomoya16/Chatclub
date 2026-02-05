package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"os"
	"sync/atomic"
	"time"
)

type HealthService interface {
	Live(ctx context.Context) bool
	Ready(ctx context.Context) bool
	Report(ctx context.Context) domain.HealthReport
	MarkReady()
	MarkNotReady()
}

type healthService struct {
	healthRepo repository.HealthRepository
	// 並行アクセスしても安全に読み書きできるbool
	readyFlag  atomic.Bool
	appVersion string
}

func NewHealthService(healthRepo repository.HealthRepository) HealthService {
	s := &healthService{
		healthRepo: healthRepo,
		appVersion: os.Getenv("APP_VERSION"),
	}
	// 起動直後はNotReady
	s.readyFlag.Store(false)
	return s
}

// 準備OK
func (s *healthService) MarkReady() {
	s.readyFlag.Store(true)
}

// 準備NO
func (s *healthService) MarkNotReady() {
	s.readyFlag.Store(false)
}

// プロセスが動いてたらOK
func (s *healthService) Live(ctx context.Context) bool {
	return true
}

// 依存到達確認
func (s *healthService) Ready(ctx context.Context) bool {
	if !s.readyFlag.Load() {
		return false
	}
	return s.healthRepo.PingDB(ctx) == nil
}

// 人間/監視向けの総合診断
func (s *healthService) Report(ctx context.Context) domain.HealthReport {
	dbOK := s.healthRepo.PingDB(ctx) == nil
	return domain.HealthReport{
		Live:    true,
		Ready:   s.readyFlag.Load() && dbOK,
		DB:      dbOK,
		Version: s.appVersion,
		Time:    time.Now().Format(time.RFC3339),
	}
}
