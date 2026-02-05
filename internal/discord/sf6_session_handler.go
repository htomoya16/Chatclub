package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) handleSF6Session(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6SessionService == nil || r.SF6Service == nil || r.SF6AccountService == nil {
		respondEphemeral(s, i, "sf6機能が無効です（Buckler設定未完）")
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

	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		followupEphemeral(s, i, "subcommandが必要です")
		return
	}
	sub := data.Options[0]

	var opponentCode, subjectCode string
	for _, opt := range sub.Options {
		switch opt.Name {
		case "opponent_code":
			opponentCode = opt.StringValue()
		case "subject_code":
			subjectCode = opt.StringValue()
		}
	}
	if opponentCode == "" {
		followupEphemeral(s, i, "opponent_code が必要です")
		return
	}

	ctx := context.Background()
	if sid, _, ok, err := r.resolveSIDFromMention(ctx, i.GuildID, opponentCode); ok {
		if err != nil {
			followupEphemeral(s, i, err.Error())
			return
		}
		opponentCode = sid
	}
	subjectSID, err := r.resolveSubjectSID(ctx, i.GuildID, userID, subjectCode)
	if err != nil {
		followupEphemeral(s, i, err.Error())
		return
	}

	switch sub.Name {
	case "start":
		startedAt := time.Now().UTC()
		_, err := r.SF6SessionService.Start(ctx, i.GuildID, userID, opponentCode, startedAt)
		if err != nil {
			followupEphemeral(s, i, "開始に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("セッション開始: subject=%s opponent=%s", subjectSID, opponentCode)
		followupEphemeral(s, i, label)
	case "end":
		endedAt := time.Now().UTC()
		session, err := r.SF6SessionService.End(ctx, i.GuildID, userID, opponentCode, endedAt)
		if err != nil {
			followupEphemeral(s, i, "終了に失敗: "+err.Error())
			return
		}
		if session == nil {
			followupEphemeral(s, i, "アクティブなセッションがありません")
			return
		}
		fetchUserID := userID
		if r.SF6AccountService != nil {
			if owner, err := r.SF6AccountService.GetByFighter(ctx, i.GuildID, subjectSID); err == nil && owner != nil && owner.UserID != "" {
				fetchUserID = owner.UserID
			}
		}
		if _, _, err := r.initialFetch(ctx, i.GuildID, fetchUserID, subjectSID); err != nil {
			followupEphemeral(s, i, "対戦記録の取得に失敗: "+err.Error())
			return
		}
		endExclusive := endedAt.Add(time.Nanosecond)
		stats, err := r.SF6Service.StatsByOpponentRange(ctx, i.GuildID, subjectSID, opponentCode, session.StartedAt, endExclusive)
		if err != nil {
			followupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("セッション: %s〜%s (JST)", formatJST(session.StartedAt), formatJST(endedAt))
		subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, i.GuildID, subjectSID, opponentCode)
		embed := buildStatsEmbed("SF6 Stats (Session)", label, subjectUser, opponentUser, stats)
		followupPublicEmbed(s, i, "", embed, nil)
	default:
		followupEphemeral(s, i, "不明なサブコマンドです")
	}
}
