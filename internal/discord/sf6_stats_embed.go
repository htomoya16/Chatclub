package discord

import (
	"backend/internal/domain"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type statsTotals struct {
	Total  int
	Wins   int
	Losses int
	Draws  int
}

type statsEmbedUser struct {
	SID     string
	UserID  string
	Mention string
	IconURL string
	DisplayName string
}

func buildStatsEmbed(title, periodLabel string, subject, opponent statsEmbedUser, rows []domain.SF6BattleStatRow) *discordgo.MessageEmbed {
	totals, byChar := summarizeStats(rows)
	winrate := calcWinRate(totals)

	desc := periodLabel
	charLines := formatCharStatsTable(byChar)
	if charLines == "" {
		charLines = "no data"
	}
	players := fmt.Sprintf("subject: %s\nopponent: %s", formatStatsUserLine(subject), formatStatsUserLine(opponent))

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       0x2b6cb0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Players",
				Value:  players,
				Inline: false,
			},
			{
				Name:   "Total",
				Value:  fmt.Sprintf("**%d**", totals.Total),
				Inline: true,
			},
			{
				Name:   "Win",
				Value:  fmt.Sprintf("**%d**", totals.Wins),
				Inline: true,
			},
			{
				Name:   "Loss",
				Value:  fmt.Sprintf("**%d**", totals.Losses),
				Inline: true,
			},
			{
				Name:   "Draw",
				Value:  fmt.Sprintf("**%d**", totals.Draws),
				Inline: true,
			},
			{
				Name:   "Win Rate (no draws)",
				Value:  fmt.Sprintf("**%s**", winrate),
				Inline: true,
			},
			{
				Name:   "By Character",
				Value:  charLines,
				Inline: false,
			},
		},
	}
	applyStatsIcons(embed, subject, opponent)
	return embed
}

func summarizeStats(rows []domain.SF6BattleStatRow) (statsTotals, map[string]statsTotals) {
	total := statsTotals{}
	byChar := make(map[string]statsTotals)
	for _, row := range rows {
		count := row.Count
		total.Total += count
		switch row.Result {
		case "win":
			total.Wins += count
		case "loss":
			total.Losses += count
		case "draw":
			total.Draws += count
		}
		char := row.SelfCharacter
		if char == "" {
			char = "unknown"
		}
		c := byChar[char]
		c.Total += count
		switch row.Result {
		case "win":
			c.Wins += count
		case "loss":
			c.Losses += count
		case "draw":
			c.Draws += count
		}
		byChar[char] = c
	}
	return total, byChar
}

func calcWinRate(t statsTotals) string {
	denom := t.Wins + t.Losses
	if denom == 0 {
		return "0.0%"
	}
	rate := float64(t.Wins) / float64(denom) * 100
	return fmt.Sprintf("%.1f%%", rate)
}

func formatCharStatsTable(byChar map[string]statsTotals) string {
	if len(byChar) == 0 {
		return ""
	}
	type entry struct {
		Char string
		Stat statsTotals
	}
	entries := make([]entry, 0, len(byChar))
	for k, v := range byChar {
		entries = append(entries, entry{Char: k, Stat: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Stat.Total > entries[j].Stat.Total
	})
	const maxLines = 8
	const maxNameWidth = 12
	lines := make([]string, 0, len(entries))
	nameWidth := 4
	for _, e := range entries {
		if w := runeLen(e.Char); w > nameWidth {
			nameWidth = w
		}
	}
	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}
	header := fmt.Sprintf("%s  %4s %4s %4s %4s %6s",
		padRight("CHAR", nameWidth), "G", "W", "L", "D", "WR")
	lines = append(lines, header)
	for i, e := range entries {
		if i >= maxLines {
			break
		}
		name := padRight(truncateName(e.Char, nameWidth), nameWidth)
		lines = append(lines, fmt.Sprintf("%s  %4d %4d %4d %4d %6s",
			name, e.Stat.Total, e.Stat.Wins, e.Stat.Losses, e.Stat.Draws, calcWinRate(e.Stat)))
	}
	if len(entries) > maxLines {
		lines = append(lines, fmt.Sprintf("...and %d more", len(entries)-maxLines))
	}
	return "```\n" + strings.Join(lines, "\n") + "\n```"
}

func runeLen(s string) int {
	return len([]rune(s))
}

func truncateName(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

func padRight(s string, width int) string {
	n := runeLen(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func formatStatsUserLine(user statsEmbedUser) string {
	name := strings.TrimSpace(user.DisplayName)
	if user.Mention != "" {
		if user.SID != "" {
			if name != "" {
				return fmt.Sprintf("%s (%s, SID: `%s`)", user.Mention, name, user.SID)
			}
			return fmt.Sprintf("%s (SID: `%s`)", user.Mention, user.SID)
		}
		return user.Mention
	}
	if user.SID != "" {
		if name != "" {
			return fmt.Sprintf("%s (SID: `%s`)", name, user.SID)
		}
		return "`" + user.SID + "`"
	}
	return "unknown"
}

func applyStatsIcons(embed *discordgo.MessageEmbed, subject, opponent statsEmbedUser) {
	if embed == nil {
		return
	}
	if subject.UserID != "" && subject.UserID == opponent.UserID {
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
	if embed.Author == nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    "Opponent",
			IconURL: opponent.IconURL,
		}
		return
	}
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: opponent.IconURL}
}
