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
	followupEphemeralEmbed(s, i, accountEmbed, sf6LinkButtons(linked))
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
	followupEphemeralEmbed(s, i, accountEmbed, sf6LinkButtons(linked))
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	user := interactionUser(i)
	if user == nil {
		return ""
	}
	return user.ID
}

func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
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

func deferEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func deferPublic(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func followupEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func (r *Router) handleSF6Component(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	data := i.MessageComponentData()
	switch data.CustomID {
	case "sf6_link_button":
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
	case "sf6_unlink_button":
		r.handleSF6Unlink(s, i)
	case "sf6_friend_add_button":
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				Title:    "SF6 Friend Add",
				CustomID: "sf6_friend_add_modal",
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
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "alias",
								Label:       "Alias (optional)",
								Style:       discordgo.TextInputShort,
								Placeholder: "覚えやすい名前",
								Required:    false,
								MaxLength:   32,
							},
						},
					},
				},
			},
		})
	case "sf6_friend_remove_button":
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				Title:    "SF6 Friend Remove",
				CustomID: "sf6_friend_remove_modal",
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
		if err := deferEphemeral(s, i); err != nil {
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
		followupPublicEmbed(s, i, "<@"+userID+"> 連携しました", accountEmbed, sf6LinkButtons(linked))
	case "sf6_friend_add_modal":
		userCode := modalValue(data.Components, "user_code")
		if userCode == "" {
			respondEphemeral(s, i, "ユーザーコードが必要")
			return
		}
		alias := modalValue(data.Components, "alias")
		userID := interactionUserID(i)
		user := interactionUser(i)
		if userID == "" {
			respondEphemeral(s, i, "user_idの取得に失敗")
			return
		}
		if r.SF6FriendService == nil {
			respondEphemeral(s, i, "sf6機能が無効です")
			return
		}
		if err := deferEphemeral(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return
		}
		displayName := ""
		if r.SF6Service != nil {
			if card, err := r.SF6Service.FetchCard(context.Background(), userCode); err == nil {
				displayName = card.FighterName
			}
		}
		friend := domain.SF6Friend{
			GuildID:     i.GuildID,
			UserID:      userID,
			FighterID:   userCode,
			DisplayName: displayName,
			Alias:       alias,
		}
		ctx := context.Background()
		if err := r.SF6FriendService.Upsert(ctx, friend); err != nil {
			followupEphemeral(s, i, "登録に失敗: "+err.Error())
			return
		}
		var fetchErr error
		if r.SF6Service != nil {
			_, _, fetchErr = r.initialFetch(ctx, i.GuildID, userID, userCode)
		}
		friends, err := r.SF6FriendService.List(ctx, i.GuildID, userID)
		if err != nil {
			followupEphemeral(s, i, "登録完了。一覧の取得に失敗: "+err.Error())
			return
		}
		embed := r.buildFriendEmbed(ctx, "SF6 Friends", i.GuildID, userID, user, friends)
		followupEphemeralEmbed(s, i, embed, sf6FriendButtons())
		if fetchErr != nil {
			followupEphemeral(s, i, "登録完了。初回取得に失敗: "+fetchErr.Error())
		}
	case "sf6_friend_remove_modal":
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
		if r.SF6FriendService == nil {
			respondEphemeral(s, i, "sf6機能が無効です")
			return
		}
		if err := deferEphemeral(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return
		}
		ctx := context.Background()
		if err := r.SF6FriendService.Delete(ctx, i.GuildID, userID, userCode); err != nil {
			followupEphemeral(s, i, "削除に失敗: "+err.Error())
			return
		}
		friends, err := r.SF6FriendService.List(ctx, i.GuildID, userID)
		if err != nil {
			followupEphemeral(s, i, "削除完了。一覧の取得に失敗: "+err.Error())
			return
		}
		embed := r.buildFriendEmbed(ctx, "SF6 Friends", i.GuildID, userID, user, friends)
		followupEphemeralEmbed(s, i, embed, sf6FriendButtons())
	default:
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
