package discord

import (
	"backend/internal/domain"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func sf6FriendButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "追加",
					Style:    discordgo.SuccessButton,
					CustomID: "sf6_friend_add_button",
				},
				discordgo.Button{
					Label:    "削除",
					Style:    discordgo.DangerButton,
					CustomID: "sf6_friend_remove_button",
				},
			},
		},
	}
}

func (r *Router) buildFriendEmbed(ctx context.Context, title, guildID, userID string, user *discordgo.User, friends []domain.SF6Friend) *discordgo.MessageEmbed {
	displayName := discordDisplayName(user)
	authorName := fmt.Sprintf("👥 %s のフレンド一覧", displayName)
	authorIcon := ""
	if user != nil && user.Avatar != "" {
		authorIcon = user.AvatarURL("")
	}

	listValue := "未登録"
	if len(friends) > 0 {
		lines := make([]string, 0, len(friends))
		for _, f := range friends {
			label := f.FighterID
			if f.DisplayName != "" {
				label += " (" + f.DisplayName + ")"
			}
			if f.Alias != "" {
				label += " [" + f.Alias + "]"
			}
			lines = append(lines, label)
		}
		listValue = strings.Join(lines, "\n")
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: "このパネルからフレンドの追加/削除ができます。",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x3498DB,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorName,
			IconURL: authorIcon,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Discord", Value: "<@" + userID + ">", Inline: true},
			{Name: "フレンド数", Value: strconv.Itoa(len(friends)), Inline: true},
			{Name: "フレンド一覧", Value: listValue, Inline: false},
		},
	}
}
