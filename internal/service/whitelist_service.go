package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/repository"
)

type WhitelistService interface {
	Add(ctx context.Context, platform, userID, note string) error
	Remove(ctx context.Context, platform, userID string) error
	IsAllowed(ctx context.Context, platform, userID string) (bool, error)
}

type whitelistService struct {
	repo repository.WhitelistRepository
}

func NewWhitelistService(r repository.WhitelistRepository) WhitelistService {
	return &whitelistService{repo: r}
}

func (s *whitelistService) Add(ctx context.Context, platform, userID, note string) error {
	return s.repo.Add(ctx, domain.Whitelist{
		Platform: platform,
		UserID:   userID,
		Note:     note,
	})
}

func (s *whitelistService) Remove(ctx context.Context, platform, userID string) error {
	return s.repo.Remove(ctx, platform, userID)
}

func (s *whitelistService) IsAllowed(ctx context.Context, platform, userID string) (bool, error) {
	return s.repo.Exists(ctx, platform, userID)
}
