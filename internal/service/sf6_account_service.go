package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"
)

type SF6AccountService interface {
	UpsertAndReassign(ctx context.Context, account domain.SF6Account) (int64, error)
}

type sf6AccountService struct {
	accountRepo repository.SF6AccountRepository
	battleRepo  repository.SF6BattleRepository
}

func NewSF6AccountService(accountRepo repository.SF6AccountRepository, battleRepo repository.SF6BattleRepository) SF6AccountService {
	return &sf6AccountService{accountRepo: accountRepo, battleRepo: battleRepo}
}

func (s *sf6AccountService) UpsertAndReassign(ctx context.Context, account domain.SF6Account) (int64, error) {
	if account.GuildID == "" || account.UserID == "" || account.FighterID == "" {
		return 0, errors.New("guildID, userID, fighterID are required")
	}
	if err := s.accountRepo.Upsert(ctx, account); err != nil {
		return 0, err
	}
	return s.battleRepo.ReassignOwnerBySubject(ctx, account.GuildID, account.FighterID, account.UserID)
}
