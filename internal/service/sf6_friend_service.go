package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"
)

type SF6FriendService interface {
	Upsert(ctx context.Context, friend domain.SF6Friend) error
	Delete(ctx context.Context, guildID, userID, fighterID string) error
	List(ctx context.Context, guildID, userID string) ([]domain.SF6Friend, error)
}

type sf6FriendService struct {
	friendRepo  repository.SF6FriendRepository
	accountRepo repository.SF6AccountRepository
	battleRepo  repository.SF6BattleRepository
}

func NewSF6FriendService(friendRepo repository.SF6FriendRepository, accountRepo repository.SF6AccountRepository, battleRepo repository.SF6BattleRepository) SF6FriendService {
	return &sf6FriendService{
		friendRepo:  friendRepo,
		accountRepo: accountRepo,
		battleRepo:  battleRepo,
	}
}

func (s *sf6FriendService) Upsert(ctx context.Context, friend domain.SF6Friend) error {
	if friend.GuildID == "" || friend.UserID == "" || friend.FighterID == "" {
		return errors.New("guildID, userID, fighterID are required")
	}
	return s.friendRepo.Upsert(ctx, friend)
}

func (s *sf6FriendService) Delete(ctx context.Context, guildID, userID, fighterID string) error {
	if guildID == "" || userID == "" || fighterID == "" {
		return errors.New("guildID, userID, fighterID are required")
	}
	if err := s.friendRepo.Delete(ctx, guildID, userID, fighterID); err != nil {
		return err
	}
	if s.accountRepo == nil || s.battleRepo == nil {
		return nil
	}
	owner, err := s.accountRepo.GetByFighter(ctx, guildID, fighterID)
	if err != nil {
		return err
	}
	if owner != nil {
		return nil
	}
	exists, err := s.friendRepo.ExistsByFighter(ctx, guildID, fighterID)
	if err != nil {
		return err
	}
	if !exists {
		_, err = s.battleRepo.MarkOwnerKindUnlinkedBySubject(ctx, guildID, fighterID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sf6FriendService) List(ctx context.Context, guildID, userID string) ([]domain.SF6Friend, error) {
	if guildID == "" || userID == "" {
		return nil, errors.New("guildID and userID are required")
	}
	return s.friendRepo.List(ctx, guildID, userID)
}
