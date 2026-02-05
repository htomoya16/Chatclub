package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) handleSF6Stats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6Service == nil || r.SF6AccountService == nil {
		respondEphemeral(s, i, "sf6機能が無効です（Buckler設定未完）")
		return
	}

	userID := interactionUserID(i)
	if userID == "" {
		respondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		respondEphemeral(s, i, "subcommandが必要です")
		return
	}
	sub := data.Options[0]

	ctx := context.Background()
	switch sub.Name {
	case "range":
		opts := sub.Options
		var opponentCode, subjectCode, fromStr, toStr string
		for _, opt := range opts {
			switch opt.Name {
			case "opponent_code":
				opponentCode = opt.StringValue()
			case "subject_code":
				subjectCode = opt.StringValue()
			case "from":
				fromStr = opt.StringValue()
			case "to":
				toStr = opt.StringValue()
			}
		}
		if opponentCode == "" || fromStr == "" || toStr == "" {
			respondEphemeral(s, i, "opponent_code/from/to が必要です")
			return
		}
		startAt, endAt, err := parseDateRangeJST(fromStr, toStr)
		if err != nil {
			respondEphemeral(s, i, "日付形式エラー: "+err.Error())
			return
		}
		if err := deferPublic(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return
		}
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
		stats, err := r.SF6Service.StatsByOpponentRange(ctx, i.GuildID, subjectSID, opponentCode, startAt, endAt)
		if err != nil {
			followupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("期間: %s〜%s (JST)", fromStr, toStr)
		subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, i.GuildID, subjectSID, opponentCode)
		embed := buildStatsEmbed("SF6 Stats (Range)", label, subjectUser, opponentUser, stats)
		followupPublicEmbed(s, i, "", embed, nil)
	case "count":
		opts := sub.Options
		var opponentCode, subjectCode string
		count := 0
		for _, opt := range opts {
			switch opt.Name {
			case "opponent_code":
				opponentCode = opt.StringValue()
			case "subject_code":
				subjectCode = opt.StringValue()
			case "count":
				count = int(opt.IntValue())
			}
		}
		if opponentCode == "" || count <= 0 {
			respondEphemeral(s, i, "opponent_code/count が必要です")
			return
		}
		if err := deferPublic(s, i); err != nil {
			respondEphemeral(s, i, "受付に失敗しました")
			return
		}
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
		stats, err := r.SF6Service.StatsByOpponentCount(ctx, i.GuildID, subjectSID, opponentCode, count)
		if err != nil {
			followupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("直近 %d 戦", count)
		subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, i.GuildID, subjectSID, opponentCode)
		embed := buildStatsEmbed("SF6 Stats (Count)", label, subjectUser, opponentUser, stats)
		followupPublicEmbed(s, i, "", embed, nil)
	default:
		respondEphemeral(s, i, "不明なサブコマンドです")
	}
}

func parseDateRangeJST(fromStr, toStr string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, err := parseDateJST(fromStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDateJST(toStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be >= from")
	}
	endExclusive := end.AddDate(0, 0, 1)
	return start, endExclusive, nil
}

func (r *Router) buildStatsEmbedUsers(ctx context.Context, s *discordgo.Session, guildID, subjectSID, opponentSID string) (statsEmbedUser, statsEmbedUser) {
	subject := r.buildStatsEmbedUser(ctx, s, guildID, subjectSID)
	opponent := r.buildStatsEmbedUser(ctx, s, guildID, opponentSID)
	return subject, opponent
}

func (r *Router) buildStatsEmbedUser(ctx context.Context, s *discordgo.Session, guildID, fighterID string) statsEmbedUser {
	info := statsEmbedUser{SID: fighterID}
	if r == nil || fighterID == "" {
		return info
	}
	if r.SF6AccountService != nil {
		account, err := r.SF6AccountService.GetByFighter(ctx, guildID, fighterID)
		if err == nil && account != nil && account.UserID != "" {
			info.UserID = account.UserID
			info.Mention = "<@" + account.UserID + ">"
			if s != nil {
				if user, err := s.User(account.UserID); err == nil && user != nil {
					info.IconURL = user.AvatarURL("")
				}
			}
		}
	}
	return r.populateStatsUserName(ctx, info)
}

func (r *Router) populateStatsUserName(ctx context.Context, info statsEmbedUser) statsEmbedUser {
	if r == nil || r.SF6Service == nil || info.SID == "" {
		return info
	}
	card, err := r.SF6Service.FetchCard(ctx, info.SID)
	if err != nil {
		return info
	}
	if card.FighterName != "" {
		info.DisplayName = card.FighterName
	}
	return info
}

func parseDateJST(value string, loc *time.Location) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("date is empty")
	}
	layouts := []string{
		"2006-01-02",
		"2006-1-2",
		"2006-01-02 15:04",
		"2006-1-2 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", value)
}
