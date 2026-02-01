package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"
	"time"
)

type SF6SessionService interface {
	Start(ctx context.Context, guildID, userID, opponentFighterID string, startedAt time.Time) (*domain.SF6Session, error)
	End(ctx context.Context, guildID, userID, opponentFighterID string, endedAt time.Time) (*domain.SF6Session, error)
	GetActive(ctx context.Context, guildID, userID, opponentFighterID string) (*domain.SF6Session, error)
}

type sf6SessionService struct {
	repo repository.SF6SessionRepository
}

func NewSF6SessionService(repo repository.SF6SessionRepository) SF6SessionService {
	return &sf6SessionService{repo: repo}
}

func (s *sf6SessionService) Start(ctx context.Context, guildID, userID, opponentFighterID string, startedAt time.Time) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	return s.repo.Start(ctx, guildID, userID, opponentFighterID, startedAt)
}

func (s *sf6SessionService) End(ctx context.Context, guildID, userID, opponentFighterID string, endedAt time.Time) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	return s.repo.End(ctx, guildID, userID, opponentFighterID, endedAt)
}

func (s *sf6SessionService) GetActive(ctx context.Context, guildID, userID, opponentFighterID string) (*domain.SF6Session, error) {
	if guildID == "" || userID == "" || opponentFighterID == "" {
		return nil, errors.New("guildID, userID, opponentFighterID are required")
	}
	return s.repo.GetActive(ctx, guildID, userID, opponentFighterID)
}
