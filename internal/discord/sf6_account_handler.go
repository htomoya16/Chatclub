package discord

import (
	"backend/internal/domain"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	userID := interactionUserID(i)
	user := interactionUser(i)
	if userID == "" {
		respondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	if err := deferEphemeral(s, i); err != nil {
		respondEphemeral(s, i, "受付に失敗しました")
		return
	}

	ctx := context.Background()
	accountEmbed, linked, err := r.buildAccountEmbed(ctx, "SF6 Account", i.GuildID, userID, user, 0, 0, 0)
	if err != nil {
		followupEphemeral(s, i, "状態取得に失敗: "+err.Error())
		return
	}
	followupEphemeralEmbed(s, i, accountEmbed, sf6LinkButtons(linked, userID))
}

func (r *Router) handleSF6Unlink(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6AccountService == nil {
		respondEphemeral(s, i, "sf6機能が無効です")
		return
	}

	userID := interactionUserID(i)
	user := interactionUser(i)
	if userID == "" {
		respondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	if err := deferEphemeral(s, i); err != nil {
		respondEphemeral(s, i, "受付に失敗しました")
		return
	}

	ctx := context.Background()
	affected, err := r.SF6AccountService.Unlink(ctx, i.GuildID, userID)
	if err != nil {
		followupEphemeral(s, i, "解除に失敗: "+err.Error())
		return
	}
	if affected == 0 {
		followupEphemeral(s, i, "未連携のため解除なし")
		return
	}
	accountEmbed, linked, err := r.buildAccountEmbed(ctx, "SF6 Account", i.GuildID, userID, user, 0, 0, 0)
	if err != nil {
		followupEphemeral(s, i, "連携解除しました")
		return
	}
	followupEphemeralEmbed(s, i, accountEmbed, sf6LinkButtons(linked, userID))
}

func (r *Router) resolveSubjectSID(ctx context.Context, guildID, userID, subjectOverride string) (string, error) {
	if r.SF6AccountService == nil {
		return "", fmt.Errorf("sf6機能が無効です")
	}
	account, err := r.SF6AccountService.GetByUser(ctx, guildID, userID)
	if err != nil {
		return "", err
	}
	if account == nil || account.FighterID == "" {
		return "", fmt.Errorf("連携アカウントが必要です")
	}
	subjectOverride = strings.TrimSpace(subjectOverride)
	if subjectOverride != "" {
		if sid, _, ok, err := r.resolveSIDFromMention(ctx, guildID, subjectOverride); ok {
			if err != nil {
				return "", err
			}
			return sid, nil
		}
		return subjectOverride, nil
	}
	return account.FighterID, nil
}

func (r *Router) resolveSIDFromMention(ctx context.Context, guildID, raw string) (string, string, bool, error) {
	userID, ok := parseUserMention(raw)
	if !ok {
		return "", "", false, nil
	}
	if r.SF6AccountService == nil {
		return "", "", true, fmt.Errorf("sf6機能が無効です")
	}
	account, err := r.SF6AccountService.GetByUser(ctx, guildID, userID)
	if err != nil {
		return "", "", true, err
	}
	if account == nil || account.FighterID == "" {
		return "", "", true, fmt.Errorf("指定ユーザーがSF6連携していません")
	}
	return account.FighterID, userID, true, nil
}

func (r *Router) handleSF6Component(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	data := i.MessageComponentData()
	switch {
	case strings.HasPrefix(data.CustomID, "sf6_link_button"):
		ownerID, ok := parseOwnedCustomID(data.CustomID, "sf6_link_button")
		if ok && ownerID != "" && interactionUserID(i) != ownerID {
			respondEphemeral(s, i, "この操作は発行者のみ実行できます")
			return
		}
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				Title:    "SF6 Link",
				CustomID: "sf6_link_modal",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "user_code",
								Label:       "SF6 user code (sid)",
								Style:       discordgo.TextInputShort,
								Placeholder: "例: 0123456789",
								Required:    true,
								MinLength:   3,
								MaxLength:   20,
							},
						},
					},
				},
			},
		})
	case strings.HasPrefix(data.CustomID, "sf6_unlink_button"):
		ownerID, ok := parseOwnedCustomID(data.CustomID, "sf6_unlink_button")
		if ok && ownerID != "" && interactionUserID(i) != ownerID {
			respondEphemeral(s, i, "この操作は発行者のみ実行できます")
			return
		}
		r.handleSF6Unlink(s, i)
	case data.CustomID == "sf6_friend_add_button" || data.CustomID == "sf6_friend_remove_button":
		r.handleSF6FriendComponent(s, i, data.CustomID)
	case strings.HasPrefix(data.CustomID, "sf6_history_page"):
		r.handleSF6HistoryComponent(s, i, data.CustomID)
	}
}

func (r *Router) handleSF6ModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	data := i.ModalSubmitData()
	if sf6DebugEnabled() {
		fmt.Printf("[sf6][modal] custom_id=%s\n", data.CustomID)
	}
	switch data.CustomID {
	case "sf6_link_modal":
		userCode := modalValue(data.Components, "user_code")
		if userCode == "" {
			respondEphemeral(s, i, "ユーザーコードが必要")
			return
		}

		userID := interactionUserID(i)
		user := interactionUser(i)
		if userID == "" {
			respondEphemeral(s, i, "user_idの取得に失敗")
			return
		}
		if r.SF6AccountService == nil {
			respondEphemeral(s, i, "sf6機能が無効です")
			return
		}
		if err := deferPublic(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return
		}

		account := domain.SF6Account{
			GuildID:   i.GuildID,
			UserID:    userID,
			FighterID: userCode,
			Status:    "active",
		}
		ctx := context.Background()
		updated, err := r.SF6AccountService.UpsertAndReassign(ctx, account)
		if err != nil {
			followupEphemeral(s, i, "登録に失敗: "+err.Error())
			return
		}
		totalSaved, pagesFetched, err := r.initialFetch(ctx, i.GuildID, userID, userCode)
		if err != nil {
			followupEphemeral(s, i, "登録完了。移し替え件数: "+strconv.FormatInt(updated, 10)+" / 初回取得に失敗: "+err.Error())
			return
		}
		accountEmbed, linked, err := r.buildAccountEmbed(ctx, "SF6 Account", i.GuildID, userID, user, totalSaved, pagesFetched, updated)
		if err != nil {
			followupEphemeral(s, i, "登録完了。移し替え件数: "+strconv.FormatInt(updated, 10))
			return
		}
		followupPublicEmbed(s, i, "<@"+userID+"> 連携しました", accountEmbed, sf6LinkButtons(linked, userID))
	default:
		if r.handleSF6FriendModalSubmit(s, i) {
			return
		}
		if strings.HasPrefix(data.CustomID, "sf6_") {
			respondEphemeral(s, i, "不明な操作です。もう一度お試しください。")
		}
		return
	}
}

func (r *Router) initialFetch(ctx context.Context, guildID, userID, userCode string) (int, int, error) {
	if r.SF6Service == nil {
		return 0, 0, nil
	}
	maxPages := 10
	if v := os.Getenv("SF6_POLL_MAX_PAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxPages = n
		}
	}
	totalSaved := 0
	pagesFetched := 0
	for p := 1; p <= maxPages; p++ {
		count, allExisting, err := r.SF6Service.FetchAndStoreCustomBattles(ctx, guildID, userID, userCode, p)
		if err != nil {
			return totalSaved, pagesFetched, err
		}
		pagesFetched++
		totalSaved += count
		if allExisting {
			break
		}
	}
	return totalSaved, pagesFetched, nil
}
