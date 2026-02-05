package discord

import (
	"backend/internal/domain"
	"context"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) handleSF6Friend(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6FriendService == nil {
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
	friends, err := r.SF6FriendService.List(ctx, i.GuildID, userID)
	if err != nil {
		followupEphemeral(s, i, "取得に失敗: "+err.Error())
		return
	}

	embed := r.buildFriendEmbed(ctx, "SF6 Friends", i.GuildID, userID, user, friends)
	followupEphemeralEmbed(s, i, embed, sf6FriendButtons())
}

func (r *Router) handleSF6FriendComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	switch customID {
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

func (r *Router) handleSF6FriendModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Type != discordgo.InteractionModalSubmit {
		return false
	}
	data := i.ModalSubmitData()
	switch data.CustomID {
	case "sf6_friend_add_modal":
		userCode := modalValue(data.Components, "user_code")
		if userCode == "" {
			respondEphemeral(s, i, "ユーザーコードが必要")
			return true
		}
		alias := modalValue(data.Components, "alias")
		userID := interactionUserID(i)
		user := interactionUser(i)
		if userID == "" {
			respondEphemeral(s, i, "user_idの取得に失敗")
			return true
		}
		if r.SF6FriendService == nil {
			respondEphemeral(s, i, "sf6機能が無効です")
			return true
		}
		if err := deferEphemeral(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return true
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
			return true
		}
		var fetchErr error
		if r.SF6Service != nil {
			_, _, fetchErr = r.initialFetch(ctx, i.GuildID, userID, userCode)
		}
		friends, err := r.SF6FriendService.List(ctx, i.GuildID, userID)
		if err != nil {
			followupEphemeral(s, i, "登録完了。一覧の取得に失敗: "+err.Error())
			return true
		}
		embed := r.buildFriendEmbed(ctx, "SF6 Friends", i.GuildID, userID, user, friends)
		followupEphemeralEmbed(s, i, embed, sf6FriendButtons())
		if fetchErr != nil {
			followupEphemeral(s, i, "登録完了。初回取得に失敗: "+fetchErr.Error())
		}
		return true
	case "sf6_friend_remove_modal":
		userCode := modalValue(data.Components, "user_code")
		if userCode == "" {
			respondEphemeral(s, i, "ユーザーコードが必要")
			return true
		}
		userID := interactionUserID(i)
		user := interactionUser(i)
		if userID == "" {
			respondEphemeral(s, i, "user_idの取得に失敗")
			return true
		}
		if r.SF6FriendService == nil {
			respondEphemeral(s, i, "sf6機能が無効です")
			return true
		}
		if err := deferEphemeral(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return true
		}
		ctx := context.Background()
		if err := r.SF6FriendService.Delete(ctx, i.GuildID, userID, userCode); err != nil {
			followupEphemeral(s, i, "削除に失敗: "+err.Error())
			return true
		}
		friends, err := r.SF6FriendService.List(ctx, i.GuildID, userID)
		if err != nil {
			followupEphemeral(s, i, "削除完了。一覧の取得に失敗: "+err.Error())
			return true
		}
		embed := r.buildFriendEmbed(ctx, "SF6 Friends", i.GuildID, userID, user, friends)
		followupEphemeralEmbed(s, i, embed, sf6FriendButtons())
		return true
	default:
		return false
	}
}
