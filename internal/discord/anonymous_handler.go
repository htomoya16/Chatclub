package discord

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/internal/domain"

	"github.com/bwmarrin/discordgo"
)

const (
	anonWebhookName = "chatclub-anon"
	anonUsername    = "anonymous"
)

func (r *Router) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Message == nil || m.Author == nil {
		return
	}
	if m.Author.Bot || m.WebhookID != "" {
		return
	}
	if m.GuildID == "" {
		return
	}
	if m.Type != discordgo.MessageTypeDefault && m.Type != discordgo.MessageTypeReply {
		return
	}
	if m.Content == "" && len(m.Attachments) == 0 {
		return
	}

	ctx := context.Background()
	ac, err := r.AnonymousChannelService.Get(ctx, m.GuildID, m.ChannelID)
	if err != nil || ac == nil {
		return
	}

	// Delete first as per spec.
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		return
	}

	params, err := buildWebhookParams(ctx, m.Content, m.Attachments)
	if err != nil {
		return
	}

	if err := r.executeAnonymousWebhook(ctx, s, m.ChannelID, ac, params); err != nil {
		return
	}
}

func (r *Router) handleAnon(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}

	var content string
	attachmentIDs := make([]string, 0, 3)
	for _, opt := range data.Options {
		switch opt.Name {
		case "message":
			content = opt.StringValue()
		case "file1", "file2", "file3":
			if opt.Value != nil {
				if id, ok := opt.Value.(string); ok && id != "" {
					attachmentIDs = append(attachmentIDs, id)
				}
			}
		}
	}

	attachments := resolveAttachments(data.Resolved, attachmentIDs)
	if content == "" && len(attachments) == 0 {
		respondEphemeral(s, i, "本文か添付のどちらかが必要")
		return
	}

	ctx := context.Background()
	params, err := buildWebhookParams(ctx, content, attachments)
	if err != nil {
		respondEphemeral(s, i, "匿名投稿の準備に失敗した")
		return
	}

	webhook, err := r.getOrCreateWebhook(s, i.ChannelID)
	if err != nil {
		respondEphemeral(s, i, "Webhook の準備に失敗した")
		return
	}

	if _, err := s.WebhookExecute(webhook.ID, webhook.Token, true, params); err != nil {
		respondEphemeral(s, i, "匿名投稿に失敗した")
		return
	}

	respondEphemeral(s, i, "投稿しました")
}

func (r *Router) handleAnonChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}

	if len(data.Options) == 0 {
		respondEphemeral(s, i, "サブコマンドが必要")
		return
	}

	sub := data.Options[0]
	if sub.Type != discordgo.ApplicationCommandOptionSubCommand {
		respondEphemeral(s, i, "サブコマンドが必要")
		return
	}

	var channelID string
	for _, opt := range sub.Options {
		if opt.Name == "channel" {
			if v, ok := opt.Value.(string); ok && v != "" {
				channelID = v
			}
			break
		}
	}
	if channelID == "" {
		respondEphemeral(s, i, "channel が必要")
		return
	}

	switch sub.Name {
	case "add":
		webhook, err := r.getOrCreateWebhook(s, channelID)
		if err != nil {
			respondEphemeral(s, i, "Webhook の準備に失敗した")
			return
		}

		ac := domain.AnonymousChannel{
			GuildID:      i.GuildID,
			ChannelID:    channelID,
			WebhookID:    webhook.ID,
			WebhookToken: webhook.Token,
		}

		if err := r.AnonymousChannelService.Upsert(context.Background(), ac); err != nil {
			respondEphemeral(s, i, "登録に失敗した")
			return
		}
		respondEphemeral(s, i, "匿名チャンネルに登録しました")
	case "remove":
		ctx := context.Background()
		ac, err := r.AnonymousChannelService.Get(ctx, i.GuildID, channelID)
		if err != nil {
			respondEphemeral(s, i, "取得に失敗した")
			return
		}
		if ac == nil {
			respondEphemeral(s, i, "対象チャンネルは未登録")
			return
		}

		if err := r.AnonymousChannelService.Delete(ctx, i.GuildID, channelID); err != nil {
			respondEphemeral(s, i, "削除に失敗した")
			return
		}
		if ac.WebhookID != "" {
			_ = s.WebhookDelete(ac.WebhookID)
		}

		respondEphemeral(s, i, "匿名チャンネルを解除しました")
	default:
		respondEphemeral(s, i, "不明なサブコマンド")
	}
}

func (r *Router) executeAnonymousWebhook(ctx context.Context, s *discordgo.Session, channelID string, ac *domain.AnonymousChannel, params *discordgo.WebhookParams) error {
	if ac == nil {
		return errors.New("anonymous channel not found")
	}

	if ac.WebhookID != "" && ac.WebhookToken != "" {
		if _, err := s.WebhookExecute(ac.WebhookID, ac.WebhookToken, true, params); err == nil {
			return nil
		}
	}

	webhook, err := r.getOrCreateWebhook(s, channelID)
	if err != nil {
		return err
	}

	if err := r.AnonymousChannelService.Upsert(ctx, domain.AnonymousChannel{
		GuildID:      ac.GuildID,
		ChannelID:    ac.ChannelID,
		WebhookID:    webhook.ID,
		WebhookToken: webhook.Token,
	}); err != nil {
		return err
	}

	_, err = s.WebhookExecute(webhook.ID, webhook.Token, true, params)
	return err
}

func (r *Router) getOrCreateWebhook(s *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	hooks, err := s.ChannelWebhooks(channelID)
	if err == nil {
		for _, h := range hooks {
			if h != nil && h.Name == anonWebhookName && h.Token != "" {
				return h, nil
			}
		}
	}
	return s.WebhookCreate(channelID, anonWebhookName, "")
}

func resolveAttachments(resolved *discordgo.ApplicationCommandInteractionDataResolved, ids []string) []*discordgo.MessageAttachment {
	if resolved == nil || len(ids) == 0 {
		return nil
	}
	out := make([]*discordgo.MessageAttachment, 0, len(ids))
	for _, id := range ids {
		if att, ok := resolved.Attachments[id]; ok {
			out = append(out, att)
		}
	}
	return out
}

func buildWebhookParams(ctx context.Context, content string, attachments []*discordgo.MessageAttachment) (*discordgo.WebhookParams, error) {
	params := &discordgo.WebhookParams{
		Content:  content,
		Username: anonUsername,
	}

	if len(attachments) == 0 {
		return params, nil
	}

	files := make([]*discordgo.File, 0, len(attachments))
	var failed int

	for _, att := range attachments {
		if att == nil || att.URL == "" {
			failed++
			continue
		}

		data, err := downloadAttachment(ctx, att.URL)
		if err != nil {
			failed++
			continue
		}
		files = append(files, &discordgo.File{
			Name:   att.Filename,
			Reader: bytes.NewReader(data),
		})
	}

	if len(files) > 0 {
		params.Files = files
	}

	if failed > 0 {
		if params.Content == "" {
			params.Content = "一部の添付は省略された"
		} else {
			params.Content = params.Content + "\n\n一部の添付は省略された"
		}
	}

	return params, nil
}

func downloadAttachment(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
