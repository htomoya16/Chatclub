package common

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func RespondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func RespondEphemeralEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func FollowupPublicEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content:    content,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
}

func FollowupEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func FollowupEphemeralEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
		Flags:      discordgo.MessageFlagsEphemeral,
	})
}

func DeferEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func DeferPublic(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func CommandContext() (context.Context, context.CancelFunc) {
	timeout := envDuration("DISCORD_COMMAND_TIMEOUT", 60*time.Second)
	return context.WithTimeout(context.Background(), timeout)
}

func CommandContextForInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (context.Context, context.CancelFunc) {
	ctx, cancel := CommandContext()
	if s == nil || i == nil {
		return ctx, cancel
	}
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			Logf("[discord][timeout] interaction timeout: id=%s guild=%s", i.ID, i.GuildID)
			NotifyCommandTimeout(s, i)
		}
	}()
	return ctx, cancel
}

func NotifyCommandTimeout(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		return
	}
	msg := "処理が中断した"
	if err := EditInteractionResponse(s, i, msg, nil, nil); err == nil {
		return
	} else {
		Logf("[discord][timeout] edit response failed: id=%s err=%v", i.ID, err)
	}
	if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	}); err == nil {
		return
	} else {
		Logf("[discord][timeout] followup failed: id=%s err=%v", i.ID, err)
	}
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		Logf("[discord][timeout] respond failed: id=%s err=%v", i.ID, err)
	}
}

func EditInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) error {
	if s == nil || i == nil {
		return fmt.Errorf("interaction not available")
	}
	edit := &discordgo.WebhookEdit{}
	if content != "" {
		edit.Content = &content
	}
	if embed != nil {
		embeds := []*discordgo.MessageEmbed{embed}
		edit.Embeds = &embeds
	}
	if components != nil {
		edit.Components = &components
	}
	_, err := s.InteractionResponseEdit(i.Interaction, edit)
	return err
}

func DebugEnabled() bool {
	if envBool("SF6_DEBUG") {
		return true
	}
	if envBool("BUCKLER_DEBUG") {
		return true
	}
	return envBool("DISCORD_DEBUG")
}

func Logf(format string, args ...any) {
	if DebugEnabled() {
		fmt.Printf(format+"\n", args...)
	}
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}

func envDuration(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return d
}

func ModalValue(components []discordgo.MessageComponent, customID string) string {
	for _, c := range components {
		var row *discordgo.ActionsRow
		switch v := c.(type) {
		case discordgo.ActionsRow:
			row = &v
		case *discordgo.ActionsRow:
			row = v
		default:
			continue
		}
		if row == nil {
			continue
		}
		for _, inner := range row.Components {
			switch input := inner.(type) {
			case discordgo.TextInput:
				if input.CustomID == customID {
					return input.Value
				}
			case *discordgo.TextInput:
				if input != nil && input.CustomID == customID {
					return input.Value
				}
			}
		}
	}
	return ""
}

func InteractionUserID(i *discordgo.InteractionCreate) string {
	user := InteractionUser(i)
	if user == nil {
		return ""
	}
	return user.ID
}

func InteractionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i == nil {
		return nil
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	if i.User != nil {
		return i.User
	}
	return nil
}

func ParseUserMention(value string) (string, bool) {
	v := strings.TrimSpace(value)
	if !strings.HasPrefix(v, "<@") || !strings.HasSuffix(v, ">") {
		return "", false
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(v, "<@"), ">")
	inner = strings.TrimPrefix(inner, "!")
	if inner == "" {
		return "", false
	}
	for _, r := range inner {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return inner, true
}

func ParseOwnedCustomID(customID, prefix string) (string, bool) {
	if customID == prefix {
		return "", true
	}
	if !strings.HasPrefix(customID, prefix+":") {
		return "", false
	}
	return strings.TrimPrefix(customID, prefix+":"), true
}

func FormatJST(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return t.Format("2006-01-02 15:04")
	}
	return t.In(loc).Format("2006-01-02 15:04")
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func DiscordDisplayName(user *discordgo.User) string {
	if user == nil {
		return "ユーザー"
	}
	if user.GlobalName != "" {
		return user.GlobalName
	}
	if user.Username != "" {
		return user.Username
	}
	return "ユーザー"
}
