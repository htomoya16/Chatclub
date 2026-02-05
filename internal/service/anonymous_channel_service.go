package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
)

type AnonymousChannelService interface {
	Upsert(ctx context.Context, ac domain.AnonymousChannel) error
	Delete(ctx context.Context, guildID, channelID string) error
	Get(ctx context.Context, guildID, channelID string) (*domain.AnonymousChannel, error)
}

type anonymousChannelService struct {
	repo repository.AnonymousChannelRepository
}

func NewAnonymousChannelService(repo repository.AnonymousChannelRepository) AnonymousChannelService {
	return &anonymousChannelService{repo: repo}
}

func (s *anonymousChannelService) Upsert(ctx context.Context, ac domain.AnonymousChannel) error {
	return s.repo.Upsert(ctx, ac)
}

func (s *anonymousChannelService) Delete(ctx context.Context, guildID, channelID string) error {
	return s.repo.Delete(ctx, guildID, channelID)
}

func (s *anonymousChannelService) Get(ctx context.Context, guildID, channelID string) (*domain.AnonymousChannel, error) {
	return s.repo.Get(ctx, guildID, channelID)
}
