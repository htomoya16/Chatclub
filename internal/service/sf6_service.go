package service

import (
	"backend/internal/buckler"
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type BucklerClient interface {
	FetchCustomBattlelog(ctx context.Context, sid string, page int) (buckler.BattlelogResponse, error)
	FetchCard(ctx context.Context, sid string) (buckler.CardResponse, error)
}

type SF6Service interface {
	FetchAndStoreCustomBattles(ctx context.Context, guildID, userID, sid string, page int) (int, bool, error)
	FetchCard(ctx context.Context, sid string) (buckler.CardResponse, error)
}

type sf6Service struct {
	bucklerClient BucklerClient
	battleRepo    repository.SF6BattleRepository
	accountRepo   repository.SF6AccountRepository
}

func NewSF6Service(bucklerClient BucklerClient, battleRepo repository.SF6BattleRepository, accountRepo repository.SF6AccountRepository) SF6Service {
	return &sf6Service{bucklerClient: bucklerClient, battleRepo: battleRepo, accountRepo: accountRepo}
}

func (s *sf6Service) FetchAndStoreCustomBattles(ctx context.Context, guildID, userID, sid string, page int) (int, bool, error) {
	if guildID == "" || userID == "" || sid == "" {
		return 0, false, errors.New("guildID, userID, sid are required")
	}
	ownerKind := "account"
	if s.accountRepo != nil {
		owner, err := s.accountRepo.GetByFighter(ctx, guildID, sid)
		if err != nil {
			return 0, false, err
		}
		if owner != nil {
			if owner.UserID != userID {
				return 0, true, nil
			}
			ownerKind = "account"
		} else {
			ownerKind = "friend"
		}
	}
	res, err := s.bucklerClient.FetchCustomBattlelog(ctx, sid, page)
	if err != nil {
		return 0, false, err
	}
	battles := make([]domain.SF6Battle, 0, len(res.PageProps.ReplayList))
	for _, entry := range res.PageProps.ReplayList {
		battle, ok := buildBattleFromReplay(guildID, userID, sid, ownerKind, entry)
		if !ok {
			continue
		}
		battles = append(battles, battle)
	}
	if len(battles) == 0 {
		return 0, true, nil
	}
	keys := make([]string, 0, len(battles))
	for _, battle := range battles {
		keys = append(keys, battle.SourceKey)
	}
	exists, err := s.battleRepo.ExistingSourceKeys(ctx, guildID, sid, keys)
	if err != nil {
		return 0, false, err
	}
	if len(exists) == len(keys) {
		return 0, true, nil
	}
	count, err := s.battleRepo.BulkUpsert(ctx, battles)
	if err != nil {
		return 0, false, err
	}
	return count, false, nil
}

func (s *sf6Service) FetchCard(ctx context.Context, sid string) (buckler.CardResponse, error) {
	if sid == "" {
		return buckler.CardResponse{}, errors.New("sid is required")
	}
	return s.bucklerClient.FetchCard(ctx, sid)
}

func buildBattleFromReplay(guildID, userID, sid, ownerKind string, entry buckler.ReplayEntry) (domain.SF6Battle, bool) {
	selfSID, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return domain.SF6Battle{}, false
	}

	p1 := entry.Player1Info
	p2 := entry.Player2Info

	var (
		self buckler.PlayerInfo
		oppo buckler.PlayerInfo
	)
	switch selfSID {
	case p1.Player.ShortID:
		self, oppo = p1, p2
	case p2.Player.ShortID:
		self, oppo = p2, p1
	default:
		return domain.SF6Battle{}, false
	}

	selfWins := buckler.RoundWins(self.RoundResults)
	oppoWins := buckler.RoundWins(oppo.RoundResults)
	result := "draw"
	if selfWins > oppoWins {
		result = "win"
	} else if selfWins < oppoWins {
		result = "loss"
	}

	battleAt := time.Unix(entry.UploadedAt, 0).UTC()
	sourceKey := entry.ReplayID
	if sourceKey == "" {
		sourceKey = fmt.Sprintf("%d:%d:%d:%s:%s:%s",
			entry.UploadedAt,
			self.Player.ShortID,
			oppo.Player.ShortID,
			self.CharacterToolName,
			oppo.CharacterToolName,
			result,
		)
	}

	raw, _ := json.Marshal(entry)

	return domain.SF6Battle{
		GuildID:           guildID,
		UserID:            userID,
		OwnerKind:         ownerKind,
		SubjectFighterID:  strconv.FormatInt(self.Player.ShortID, 10),
		OpponentFighterID: strconv.FormatInt(oppo.Player.ShortID, 10),
		BattleAt:          battleAt,
		Result:            result,
		SelfCharacter:     self.CharacterToolName,
		OpponentCharacter: oppo.CharacterToolName,
		RoundWins:         selfWins,
		RoundLosses:       oppoWins,
		SourceKey:         sourceKey,
		SessionID:         nil,
		RawPayload:        raw,
	}, true
}
