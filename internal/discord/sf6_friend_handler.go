package discord

import (
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
