package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// /ping コマンドの処理。
// WhitelistService で許可ユーザかどうかを見て返答を変える。
func (r *Router) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// ユーザID取得（Guild / DM 両対応）
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	// Whitelist 未設定なら普通に pong 返すだけでも良いが、
	// 今回はちゃんとチェックする。
	ctx := context.Background()
	ok, err := r.WhitelistService.IsAllowed(ctx, "discord", userID)
	if err != nil {
		// 内部エラー時：ユーザ向けには軽いメッセージだけ返す
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "内部エラーが発生したため、現在ステータスを確認できない。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	msg := "pong"
	flags := discordgo.MessageFlags(0)

	if !ok {
		msg = "許可されていないユーザである。運営に問い合わせること。"
		flags = discordgo.MessageFlagsEphemeral
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	})
}
