package service

import (
	"context"
	"math/rand"
	"time"

	"backend/internal/repository"
)

// PollLogger defines the minimal logger interface used by the poller.
type PollLogger interface {
	Infof(format string, args ...interface{})
	Error(args ...interface{})
}

func RunSF6Poller(
	ctx context.Context,
	interval time.Duration,
	maxPages int,
	accountDelayMax time.Duration,
	accountRepo repository.SF6AccountRepository,
	sf6Service SF6Service,
	logger PollLogger,
) {
	if interval <= 0 || maxPages <= 0 {
		return
	}
	logger.Infof("sf6 poller start: interval=%s max_pages=%d account_delay_max=%s", interval, maxPages, accountDelayMax)
	runSF6PollOnce(ctx, maxPages, accountDelayMax, accountRepo, sf6Service, logger)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runSF6PollOnce(ctx, maxPages, accountDelayMax, accountRepo, sf6Service, logger)
		}
	}
}

func runSF6PollOnce(
	ctx context.Context,
	maxPages int,
	accountDelayMax time.Duration,
	accountRepo repository.SF6AccountRepository,
	sf6Service SF6Service,
	logger PollLogger,
) {
	accounts, err := accountRepo.ListActive(ctx)
	if err != nil {
		logger.Error("sf6 poll list accounts: ", err)
		return
	}
	if len(accounts) == 0 {
		logger.Infof("sf6 poll: no active accounts")
		return
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, account := range accounts {
		totalSaved := 0
		for page := 1; page <= maxPages; page++ {
			count, allExisting, err := sf6Service.FetchAndStoreCustomBattles(
				ctx,
				account.GuildID,
				account.UserID,
				account.FighterID,
				page,
			)
			if err != nil {
				logger.Error("sf6 poll fetch: ", err)
				break
			}
			totalSaved += count
			if allExisting {
				break
			}
		}
		logger.Infof("sf6 poll done: guild=%s user=%s saved=%d", account.GuildID, account.UserID, totalSaved)
		jitterSleep(ctx, rng, accountDelayMax)
	}
}

func jitterSleep(ctx context.Context, rng *rand.Rand, max time.Duration) {
	if max <= 0 {
		return
	}
	delay := time.Duration(rng.Int63n(int64(max)))
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
