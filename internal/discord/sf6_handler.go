package discord

import (
	"backend/internal/domain"
	"context"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) handleSF6Link(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6AccountService == nil {
		respondEphemeral(s, i, "sf6機能が無効です")
		return
	}

	data := i.ApplicationCommandData()
	var userCode string
	for _, opt := range data.Options {
		switch opt.Name {
		case "user_code":
			userCode = opt.StringValue()
		}
	}
	if userCode == "" {
		respondEphemeral(s, i, "ユーザーコードが必要")
		return
	}

	userID := interactionUserID(i)
	if userID == "" {
		respondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	if err := deferEphemeral(s, i); err != nil {
		respondEphemeral(s, i, "受付に失敗しました")
		return
	}

	account := domain.SF6Account{
		GuildID:     i.GuildID,
		UserID:      userID,
		FighterID:   userCode,
		Status:      "active",
	}
	ctx := context.Background()
	updated, err := r.SF6AccountService.UpsertAndReassign(ctx, account)
	if err != nil {
		followupEphemeral(s, i, "登録に失敗: "+err.Error())
		return
	}

	totalSaved := 0
	pagesFetched := 0
	if r.SF6Service != nil {
		maxPages := 10
		if v := os.Getenv("SF6_POLL_MAX_PAGES"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxPages = n
			}
		}
		for p := 1; p <= maxPages; p++ {
			count, allExisting, err := r.SF6Service.FetchAndStoreCustomBattles(ctx, i.GuildID, userID, userCode, p)
			if err != nil {
				followupEphemeral(s, i, "登録完了。移し替え件数: "+strconv.FormatInt(updated, 10)+" / 初回取得に失敗: "+err.Error())
				return
			}
			pagesFetched++
			totalSaved += count
			if allExisting {
				break
			}
		}
	}

	msg := "登録完了。移し替え件数: " + strconv.FormatInt(updated, 10)
	if pagesFetched > 0 {
		msg += " / 初回取得: " + strconv.Itoa(totalSaved) + "件 (" + strconv.Itoa(pagesFetched) + " pages)"
	}
	followupEphemeral(s, i, msg)
}

func (r *Router) handleSF6Fetch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6Service == nil {
		respondEphemeral(s, i, "sf6機能が無効です（Buckler設定未完）")
		return
	}

	data := i.ApplicationCommandData()
	var userCode string
	page := 1
	maxPages := 10
	for _, opt := range data.Options {
		switch opt.Name {
		case "user_code":
			userCode = opt.StringValue()
		case "page":
			page = int(opt.IntValue())
		}
	}
	if userCode == "" {
		respondEphemeral(s, i, "ユーザーコードが必要")
		return
	}
	if page <= 0 {
		page = 1
	}

	userID := interactionUserID(i)
	if userID == "" {
		respondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	if err := deferEphemeral(s, i); err != nil {
		respondEphemeral(s, i, "受付に失敗しました")
		return
	}

	ctx := context.Background()
	totalSaved := 0
	pagesFetched := 0
	for p := page; p < page+maxPages; p++ {
		count, allExisting, err := r.SF6Service.FetchAndStoreCustomBattles(ctx, i.GuildID, userID, userCode, p)
		if err != nil {
			followupEphemeral(s, i, "取得に失敗: "+err.Error())
			return
		}
		pagesFetched++
		totalSaved += count
	if allExisting {
			break
		}
	}
	followupEphemeral(s, i, "取得完了。保存件数: "+strconv.Itoa(totalSaved)+" / pages: "+strconv.Itoa(pagesFetched))
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func deferEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func followupEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}
