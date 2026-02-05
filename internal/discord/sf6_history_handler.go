package discord

import (
	"backend/internal/domain"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const sf6HistoryPageSize = 10

type sf6HistoryUser struct {
	SID         string
	Mention     string
	DisplayName string
	IconURL     string
}

func (r *Router) handleSF6History(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	opts := data.Options
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
		respondEphemeral(s, i, "opponent_code が必要です")
		return
	}

	ctx := context.Background()
	if sid, _, ok, err := r.resolveSIDFromMention(ctx, i.GuildID, opponentCode); ok {
		if err != nil {
			respondEphemeral(s, i, err.Error())
			return
		}
		opponentCode = sid
	}
	subjectSID, err := r.resolveSubjectSID(ctx, i.GuildID, userID, subjectCode)
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}

	if err := deferPublic(s, i); err != nil {
		respondEphemeral(s, i, "受付に失敗しました")
		return
	}

	embed, components, err := r.buildSF6HistoryEmbed(ctx, s, i.GuildID, userID, subjectSID, opponentCode, 1)
	if err != nil {
		followupEphemeral(s, i, "履歴取得に失敗: "+err.Error())
		return
	}
	followupPublicEmbed(s, i, "", embed, components)
}

func (r *Router) handleSF6HistoryComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	ownerID, subjectSID, opponentSID, page, ok := parseSF6HistoryCustomID(customID)
	if !ok {
		respondEphemeral(s, i, "不正な操作です")
		return
	}
	if ownerID != "" && interactionUserID(i) != ownerID {
		respondEphemeral(s, i, "この操作は発行者のみ実行できます")
		return
	}
	ctx := context.Background()
	embed, components, err := r.buildSF6HistoryEmbed(ctx, s, i.GuildID, ownerID, subjectSID, opponentSID, page)
	if err != nil {
		respondEphemeral(s, i, "履歴取得に失敗: "+err.Error())
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

func (r *Router) buildSF6HistoryEmbed(ctx context.Context, s *discordgo.Session, guildID, ownerID, subjectSID, opponentSID string, page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	if page <= 0 {
		page = 1
	}
	total, err := r.SF6Service.CountByOpponent(ctx, guildID, subjectSID, opponentSID)
	if err != nil {
		return nil, nil, err
	}
	totalPages := (total + sf6HistoryPageSize - 1) / sf6HistoryPageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * sf6HistoryPageSize
	rows, err := r.SF6Service.HistoryByOpponent(ctx, guildID, subjectSID, opponentSID, sf6HistoryPageSize, offset)
	if err != nil {
		return nil, nil, err
	}

	subject := r.buildSF6HistoryUser(ctx, s, guildID, subjectSID)
	opponent := r.buildSF6HistoryUser(ctx, s, guildID, opponentSID)
	desc := buildSF6HistoryDescription(subject, opponent, rows)

	embed := &discordgo.MessageEmbed{
		Title:       "SF6 History",
		Description: desc,
		Color:       0x2b6cb0,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d • Total %d", page, totalPages, total),
		},
	}
	applySF6HistoryIcons(embed, subject, opponent)

	components := buildSF6HistoryButtons(ownerID, subjectSID, opponentSID, page, totalPages)
	return embed, components, nil
}

func buildSF6HistoryDescription(subject, opponent sf6HistoryUser, rows []domain.SF6BattleHistoryRow) string {
	if len(rows) == 0 {
		return "該当データなし"
	}
	leftName := formatSF6HistoryUser(subject)
	rightName := formatSF6HistoryUser(opponent)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		leftResult, rightResult := formatSF6HistoryResult(row.Result)
		leftChar := formatSF6Character(row.SelfCharacter)
		rightChar := formatSF6Character(row.OpponentCharacter)
		line := fmt.Sprintf("%s JST\n%s [%s]  %s vs %s  [%s] %s",
			formatJST(row.BattleAt),
			leftName, leftChar, leftResult, rightResult, rightChar, rightName,
		)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n\n")
}

func (r *Router) buildSF6HistoryUser(ctx context.Context, s *discordgo.Session, guildID, fighterID string) sf6HistoryUser {
	info := sf6HistoryUser{SID: fighterID}
	if fighterID == "" {
		return info
	}
	if r != nil && r.SF6AccountService != nil {
		if account, err := r.SF6AccountService.GetByFighter(ctx, guildID, fighterID); err == nil && account != nil && account.UserID != "" {
			info.Mention = "<@" + account.UserID + ">"
			if s != nil {
				if user, err := s.User(account.UserID); err == nil && user != nil {
					info.IconURL = user.AvatarURL("")
				}
			}
		}
	}
	if r != nil && r.SF6Service != nil {
		if card, err := r.SF6Service.FetchCard(ctx, fighterID); err == nil {
			if card.FighterName != "" {
				info.DisplayName = card.FighterName
			}
		}
	}
	return info
}

func formatSF6HistoryUser(user sf6HistoryUser) string {
	if user.Mention != "" {
		return user.Mention
	}
	if user.DisplayName != "" {
		return user.DisplayName
	}
	if user.SID != "" {
		return "`" + user.SID + "`"
	}
	return "unknown"
}

func applySF6HistoryIcons(embed *discordgo.MessageEmbed, subject, opponent sf6HistoryUser) {
	if embed == nil {
		return
	}
	if subject.Mention != "" && subject.Mention == opponent.Mention {
		opponent.IconURL = ""
	}
	if subject.IconURL != "" {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    "Subject",
			IconURL: subject.IconURL,
		}
	}
	if opponent.IconURL == "" {
		return
	}
	footerText := ""
	if embed.Footer != nil {
		footerText = embed.Footer.Text
	}
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    footerText,
		IconURL: opponent.IconURL,
	}
}

func formatSF6HistoryResult(result string) (string, string) {
	switch result {
	case "win":
		return "🟩 **WIN**", "🟥 **LOSE**"
	case "loss":
		return "🟥 **LOSE**", "🟩 **WIN**"
	case "draw":
		return "🟨 **DRAW**", "🟨 **DRAW**"
	default:
		upper := strings.ToUpper(result)
		return "**" + upper + "**", "**" + upper + "**"
	}
}

func formatSF6Character(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "-"
	}
	return strings.ToUpper(name)
}

func buildSF6HistoryButtons(ownerID, subjectSID, opponentSID string, page, totalPages int) []discordgo.MessageComponent {
	prevPage := page - 1
	nextPage := page + 1
	prevDisabled := page <= 1
	nextDisabled := page >= totalPages
	if prevPage < 1 {
		prevPage = 1
	}
	if nextPage > totalPages {
		nextPage = totalPages
	}
	prevID := buildSF6HistoryCustomID(ownerID, subjectSID, opponentSID, prevPage)
	nextID := buildSF6HistoryCustomID(ownerID, subjectSID, opponentSID, nextPage)
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "前へ",
					Style:    discordgo.SecondaryButton,
					CustomID: prevID,
					Disabled: prevDisabled,
				},
				discordgo.Button{
					Label:    "次へ",
					Style:    discordgo.SecondaryButton,
					CustomID: nextID,
					Disabled: nextDisabled,
				},
			},
		},
	}
}

func buildSF6HistoryCustomID(ownerID, subjectSID, opponentSID string, page int) string {
	if page <= 0 {
		page = 1
	}
	return "sf6_history_page:" + ownerID + ":" + subjectSID + ":" + opponentSID + ":" + strconv.Itoa(page)
}

func parseSF6HistoryCustomID(customID string) (string, string, string, int, bool) {
	parts := strings.Split(customID, ":")
	if len(parts) != 5 {
		return "", "", "", 0, false
	}
	if parts[0] != "sf6_history_page" {
		return "", "", "", 0, false
	}
	page, err := strconv.Atoi(parts[4])
	if err != nil || page <= 0 {
		return "", "", "", 0, false
	}
	return parts[1], parts[2], parts[3], page, true
}
