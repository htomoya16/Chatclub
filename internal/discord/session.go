package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Session interface {
	Start(ctx context.Context) error
	Close() error
	AddHandler(handler any)
	RegisterCommands(ctx context.Context, appID, guildID string) error
}

type session struct {
	dg      *discordgo.Session
	started bool
}

// 固定で使うIntent。
const defaultIntents = discordgo.IntentsGuilds |
	discordgo.IntentsGuildMessages

func NewSession(token string) (Session, error) {
	if token == "" {
		return nil, fmt.Errorf("discord token is empty")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	dg.Identify.Intents = defaultIntents

	return &session{dg: dg}, nil
}

func (s *session) Start(ctx context.Context) error {
	if s.started {
		return nil
	}

	if err := s.dg.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	// 軽く待って安定させる（好みで削ってOK）
	select {
	case <-ctx.Done():
		_ = s.dg.Close()
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	s.started = true
	return nil
}

func (s *session) Close() error {
	if !s.started {
		return nil
	}
	s.started = false
	return s.dg.Close()
}

func (s *session) AddHandler(handler any) {
	s.dg.AddHandler(handler)
}

func (s *session) RegisterCommands(ctx context.Context, appID, guildID string) error {
	if appID == "" {
		return fmt.Errorf("discord app id is empty")
	}

	cmds := Commands()

	for _, cmd := range cmds {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := s.dg.ApplicationCommandCreate(appID, guildID, cmd); err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
	}

	return nil
}
