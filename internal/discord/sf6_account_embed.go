package discord

import (
	"backend/internal/buckler"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	sf6ColorLinked   = 0x2ECC71
	sf6ColorUnlinked = 0xE74C3C
)

func sf6LinkButtons(linked bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "連携",
					Style:    discordgo.PrimaryButton,
					CustomID: "sf6_link_button",
				},
				discordgo.Button{
					Label:    "解除",
					Style:    discordgo.DangerButton,
					CustomID: "sf6_unlink_button",
					Disabled: !linked,
				},
			},
		},
	}
}

func (r *Router) buildAccountEmbed(ctx context.Context, title, guildID, userID string, user *discordgo.User, totalSaved, pagesFetched int, reassigned int64) (*discordgo.MessageEmbed, bool, error) {
	account, err := r.SF6AccountService.GetByUser(ctx, guildID, userID)
	if err != nil {
		return nil, false, err
	}
	status := "❌ 未連携"
	userCode := "-"
	fighterName := "-"
	favoriteChar := ""
	linked := false
	if account != nil {
		status = "✅ 連携済み"
		userCode = account.FighterID
		linked = true
		fighterName = "取得失敗"
	}
	color := sf6ColorUnlinked
	if linked {
		color = sf6ColorLinked
	}
	displayName := discordDisplayName(user)
	statusBadge := "❌"
	if linked {
		statusBadge = "✅"
	}
	var card *buckler.CardResponse
	if linked && r.SF6Service != nil {
		c, err := r.SF6Service.FetchCard(ctx, userCode)
		if err != nil {
			if sf6DebugEnabled() {
				fmt.Printf("[sf6][embed] card fetch failed: user=%s sid=%s err=%v\n", userID, userCode, err)
			}
		} else {
			card = &c
		}
	}
	if card != nil {
		if card.FighterName != "" {
			fighterName = card.FighterName
		}
		favoriteChar = strings.ToLower(strings.TrimSpace(card.FavoriteCharacterTool))
		if favoriteChar == "common" {
			favoriteChar = ""
		}
	}
	if favoriteChar != "" {
		imageURL := characterImageURL(favoriteChar)
		if !imageExists(ctx, imageURL) {
			favoriteChar = ""
		}
	}

	mention := "<@" + userID + ">"
	authorName := fmt.Sprintf("%s %s の連携状況", statusBadge, displayName)
	authorIcon := ""
	if user != nil {
		if user.Avatar != "" {
			authorIcon = user.AvatarURL("")
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: "このパネルからSF6アカウントの連携/解除ができます。",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorName,
			IconURL: authorIcon,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ステータス", Value: status, Inline: true},
			{Name: "Discord", Value: mention, Inline: true},
			{Name: "ユーザーコード", Value: userCode, Inline: true},
			{Name: "SF6アカウント", Value: fighterName, Inline: true},
		},
	}
	charLabel := favoriteChar
	if charLabel == "" {
		charLabel = "未取得"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "使用キャラ",
		Value:  charLabel,
		Inline: true,
	})
	if favoriteChar != "" {
		imageURL := characterImageURL(favoriteChar)
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] character image: user=%s sid=%s tool=%s url=%s\n", userID, userCode, favoriteChar, imageURL)
		}
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: imageURL}
	}
	if reassigned > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "移し替え",
			Value:  strconv.FormatInt(reassigned, 10),
			Inline: true,
		})
	}
	if pagesFetched > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "初回取得",
			Value:  strconv.Itoa(totalSaved) + "件 / " + strconv.Itoa(pagesFetched) + " pages",
			Inline: true,
		})
	}
	return embed, linked, nil
}
