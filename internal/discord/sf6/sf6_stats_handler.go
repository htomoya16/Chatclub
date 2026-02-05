package sf6

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/internal/discord/common"

	"github.com/bwmarrin/discordgo"
)

func (r *Handler) handleSF6Stats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		common.RespondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6Service == nil || r.SF6AccountService == nil {
		common.RespondEphemeral(s, i, "sf6機能が無効です（Buckler設定未完）")
		return
	}

	userID := common.InteractionUserID(i)
	if userID == "" {
		common.RespondEphemeral(s, i, "user_idの取得に失敗")
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		common.RespondEphemeral(s, i, "subcommandが必要です")
		return
	}
	sub := data.Options[0]
	common.Logf("[sf6][stats] start sub=%s guild=%s user=%s", sub.Name, i.GuildID, userID)

	ctx, cancel := common.CommandContextForInteraction(s, i)
	defer cancel()
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
			common.RespondEphemeral(s, i, "opponent_code/from/to が必要です")
			return
		}
		startAt, endAt, err := parseDateRangeJST(fromStr, toStr)
		if err != nil {
			common.RespondEphemeral(s, i, "日付形式エラー: "+err.Error())
			return
		}
		if err := common.DeferPublic(s, i); err != nil {
			common.RespondEphemeral(s, i, "受付に失敗しました")
			return
		}
		if sid, _, ok, err := r.resolveSIDFromMention(ctx, i.GuildID, opponentCode); ok {
			if err != nil {
				common.FollowupEphemeral(s, i, err.Error())
				return
			}
			opponentCode = sid
		}
		subjectSID, err := r.resolveSubjectSID(ctx, i.GuildID, userID, subjectCode)
		if err != nil {
			common.FollowupEphemeral(s, i, err.Error())
			return
		}
		if err := r.fetchLatestForStats(ctx, i.GuildID, userID, subjectSID); err != nil {
			common.FollowupEphemeral(s, i, "最新取得に失敗: "+err.Error())
			return
		}
		stats, err := r.SF6Service.StatsByOpponentRange(ctx, i.GuildID, subjectSID, opponentCode, startAt, endAt)
		if err != nil {
			common.FollowupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("期間: %s〜%s (JST)", fromStr, toStr)
		subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, i.GuildID, subjectSID, opponentCode)
		embed := buildStatsEmbed("SF6 Stats (Range)", label, subjectUser, opponentUser, stats)
		common.FollowupPublicEmbed(s, i, "", embed, nil)
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
			common.RespondEphemeral(s, i, "opponent_code/count が必要です")
			return
		}
		if err := common.DeferPublic(s, i); err != nil {
			common.RespondEphemeral(s, i, "受付に失敗しました")
			return
		}
		if sid, _, ok, err := r.resolveSIDFromMention(ctx, i.GuildID, opponentCode); ok {
			if err != nil {
				common.FollowupEphemeral(s, i, err.Error())
				return
			}
			opponentCode = sid
		}
		subjectSID, err := r.resolveSubjectSID(ctx, i.GuildID, userID, subjectCode)
		if err != nil {
			common.FollowupEphemeral(s, i, err.Error())
			return
		}
		if err := r.fetchLatestForStats(ctx, i.GuildID, userID, subjectSID); err != nil {
			common.FollowupEphemeral(s, i, "最新取得に失敗: "+err.Error())
			return
		}
		stats, err := r.SF6Service.StatsByOpponentCount(ctx, i.GuildID, subjectSID, opponentCode, count)
		if err != nil {
			common.FollowupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		label := fmt.Sprintf("直近 %d 戦", count)
		subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, i.GuildID, subjectSID, opponentCode)
		embed := buildStatsEmbed("SF6 Stats (Count)", label, subjectUser, opponentUser, stats)
		common.FollowupPublicEmbed(s, i, "", embed, nil)
	case "set":
		opts := sub.Options
		var opponentCode, subjectCode string
		for _, opt := range opts {
			switch opt.Name {
			case "opponent_code":
				opponentCode = opt.StringValue()
			case "subject_code":
				subjectCode = opt.StringValue()
			}
		}
		if opponentCode == "" {
			common.RespondEphemeral(s, i, "opponent_code が必要です")
			return
		}
		common.Logf("[sf6][stats-set] params guild=%s user=%s subject=%s opponent=%s", i.GuildID, userID, subjectCode, opponentCode)
		if err := common.DeferPublic(s, i); err != nil {
			common.RespondEphemeral(s, i, "受付に失敗しました")
			return
		}
		common.Logf("[sf6][stats-set] deferred guild=%s user=%s", i.GuildID, userID)
		if sid, _, ok, err := r.resolveSIDFromMention(ctx, i.GuildID, opponentCode); ok {
			if err != nil {
				common.FollowupEphemeral(s, i, err.Error())
				return
			}
			opponentCode = sid
		}
		subjectSID, err := r.resolveSubjectSID(ctx, i.GuildID, userID, subjectCode)
		if err != nil {
			common.FollowupEphemeral(s, i, err.Error())
			return
		}
		if err := r.fetchLatestForStats(ctx, i.GuildID, userID, subjectSID); err != nil {
			_ = common.EditInteractionResponse(s, i, "最新取得に失敗しました", nil, nil)
			common.FollowupEphemeral(s, i, "最新取得に失敗: "+err.Error())
			return
		}
		common.Logf("[sf6][stats-set] fetch ok guild=%s user=%s subject=%s opponent=%s", i.GuildID, userID, subjectSID, opponentCode)
		embed, components, err := r.buildSF6StatsSetEmbed(ctx, s, i.GuildID, userID, subjectSID, opponentCode, 1)
		if err != nil {
			_ = common.EditInteractionResponse(s, i, "集計に失敗しました", nil, nil)
			common.FollowupEphemeral(s, i, "集計に失敗: "+err.Error())
			return
		}
		if err := common.EditInteractionResponse(s, i, "", embed, components); err != nil {
			common.Logf("[sf6][stats-set] edit response failed: guild=%s user=%s err=%v", i.GuildID, userID, err)
			common.FollowupPublicEmbed(s, i, "", embed, components)
		} else {
			common.Logf("[sf6][stats-set] response updated guild=%s user=%s", i.GuildID, userID)
		}
	default:
		common.RespondEphemeral(s, i, "不明なサブコマンドです")
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

func (r *Handler) buildStatsEmbedUsers(ctx context.Context, s *discordgo.Session, guildID, subjectSID, opponentSID string) (statsEmbedUser, statsEmbedUser) {
	subject := r.buildStatsEmbedUser(ctx, s, guildID, subjectSID)
	opponent := r.buildStatsEmbedUser(ctx, s, guildID, opponentSID)
	return subject, opponent
}

func (r *Handler) buildStatsEmbedUser(ctx context.Context, s *discordgo.Session, guildID, fighterID string) statsEmbedUser {
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

func (r *Handler) populateStatsUserName(ctx context.Context, info statsEmbedUser) statsEmbedUser {
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

func (r *Handler) fetchLatestForStats(ctx context.Context, guildID, requestUserID, subjectSID string) error {
	if r == nil || r.SF6Service == nil {
		return nil
	}
	fetchUserID := requestUserID
	if r.SF6AccountService != nil {
		if owner, err := r.SF6AccountService.GetByFighter(ctx, guildID, subjectSID); err == nil && owner != nil && owner.UserID != "" {
			fetchUserID = owner.UserID
		}
	}
	_, _, err := r.initialFetch(ctx, guildID, fetchUserID, subjectSID)
	return err
}

func (r *Handler) handleSF6StatsSetComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	ownerID, subjectSID, opponentSID, page, ok := parseSF6StatsSetCustomID(customID)
	if !ok {
		common.RespondEphemeral(s, i, "不正な操作です")
		return
	}
	if ownerID != "" && common.InteractionUserID(i) != ownerID {
		common.RespondEphemeral(s, i, "この操作は発行者のみ実行できます")
		return
	}
	ctx, cancel := common.CommandContextForInteraction(s, i)
	defer cancel()
	embed, components, err := r.buildSF6StatsSetEmbed(ctx, s, i.GuildID, ownerID, subjectSID, opponentSID, page)
	if err != nil {
		common.RespondEphemeral(s, i, "集計に失敗: "+err.Error())
		return
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

type statsSetGroup struct {
	Start time.Time
	End   time.Time
	Count int
}

func (r *Handler) buildSF6StatsSetEmbed(ctx context.Context, s *discordgo.Session, guildID, ownerID, subjectSID, opponentSID string, page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	if page <= 0 {
		page = 1
	}
	times, err := r.SF6Service.BattleTimesByOpponent(ctx, guildID, subjectSID, opponentSID)
	if err != nil {
		return nil, nil, err
	}
	groups := groupStatsSet(times, 30*time.Minute)
	if len(groups) == 0 {
		return nil, nil, fmt.Errorf("該当データなし")
	}
	totalPages := len(groups)
	if page > totalPages {
		page = totalPages
	}
	group := groups[page-1]
	endExclusive := group.End.Add(time.Nanosecond)
	stats, err := r.SF6Service.StatsByOpponentRange(ctx, guildID, subjectSID, opponentSID, group.Start, endExclusive)
	if err != nil {
		return nil, nil, err
	}
	label := fmt.Sprintf("期間: %s〜%s (JST) / %d戦", formatJST(group.Start), formatJST(group.End), group.Count)
	subjectUser, opponentUser := r.buildStatsEmbedUsers(ctx, s, guildID, subjectSID, opponentSID)
	embed := buildStatsEmbed("SF6 Stats (Set)", label, subjectUser, opponentUser, stats)
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Set %d/%d • gap<=30m", page, totalPages),
	}
	components := buildSF6StatsSetButtons(ownerID, subjectSID, opponentSID, page, totalPages)
	return embed, components, nil
}

func groupStatsSet(times []time.Time, gap time.Duration) []statsSetGroup {
	if len(times) == 0 {
		return nil
	}
	groups := make([]statsSetGroup, 0, len(times))
	current := statsSetGroup{
		Start: times[0],
		End:   times[0],
		Count: 1,
	}
	prev := times[0]
	for i := 1; i < len(times); i++ {
		t := times[i]
		if prev.Sub(t) <= gap {
			current.Start = t
			current.Count++
		} else {
			groups = append(groups, current)
			current = statsSetGroup{
				Start: t,
				End:   t,
				Count: 1,
			}
		}
		prev = t
	}
	groups = append(groups, current)
	return groups
}

func buildSF6StatsSetButtons(ownerID, subjectSID, opponentSID string, page, totalPages int) []discordgo.MessageComponent {
	firstPage := 1
	lastPage := totalPages
	prevPage := page - 1
	nextPage := page + 1
	prevDisabled := page <= 1
	nextDisabled := page >= totalPages
	firstID := buildSF6StatsSetCustomIDWithAction("first", ownerID, subjectSID, opponentSID, firstPage)
	prevID := buildSF6StatsSetCustomIDWithAction("prev", ownerID, subjectSID, opponentSID, maxInt(prevPage, 1))
	nextID := buildSF6StatsSetCustomIDWithAction("next", ownerID, subjectSID, opponentSID, minInt(nextPage, totalPages))
	lastID := buildSF6StatsSetCustomIDWithAction("last", ownerID, subjectSID, opponentSID, lastPage)
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "⏮ 最初",
					Style:    discordgo.SecondaryButton,
					CustomID: firstID,
					Disabled: prevDisabled,
				},
				discordgo.Button{
					Label:    "◀ 前へ",
					Style:    discordgo.SecondaryButton,
					CustomID: prevID,
					Disabled: prevDisabled,
				},
				discordgo.Button{
					Label:    "次へ ▶",
					Style:    discordgo.SecondaryButton,
					CustomID: nextID,
					Disabled: nextDisabled,
				},
				discordgo.Button{
					Label:    "最後 ⏭",
					Style:    discordgo.SecondaryButton,
					CustomID: lastID,
					Disabled: nextDisabled,
				},
			},
		},
	}
}

func buildSF6StatsSetCustomID(ownerID, subjectSID, opponentSID string, page int) string {
	return buildSF6StatsSetCustomIDWithAction("page", ownerID, subjectSID, opponentSID, page)
}

func buildSF6StatsSetCustomIDWithAction(action, ownerID, subjectSID, opponentSID string, page int) string {
	if page <= 0 {
		page = 1
	}
	return "sf6_stats_set_page:" + action + ":" + ownerID + ":" + subjectSID + ":" + opponentSID + ":" + strconv.Itoa(page)
}

func parseSF6StatsSetCustomID(customID string) (string, string, string, int, bool) {
	parts := strings.Split(customID, ":")
	if len(parts) != 5 && len(parts) != 6 {
		return "", "", "", 0, false
	}
	if parts[0] != "sf6_stats_set_page" {
		return "", "", "", 0, false
	}
	if len(parts) == 5 {
		page, err := strconv.Atoi(parts[4])
		if err != nil || page <= 0 {
			return "", "", "", 0, false
		}
		return parts[1], parts[2], parts[3], page, true
	}
	page, err := strconv.Atoi(parts[5])
	if err != nil || page <= 0 {
		return "", "", "", 0, false
	}
	return parts[2], parts[3], parts[4], page, true
}
