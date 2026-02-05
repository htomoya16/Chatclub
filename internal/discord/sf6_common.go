package discord

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func followupEphemeralEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
		Flags:      discordgo.MessageFlagsEphemeral,
	})
}

func respondEphemeralEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func followupPublicEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content:    content,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
}

func discordDisplayName(user *discordgo.User) string {
	if user == nil {
		return "ユーザー"
	}
	if user.GlobalName != "" {
		return user.GlobalName
	}
	if user.Username != "" {
		return user.Username
	}
	return "ユーザー"
}

func characterImageURL(toolName string) string {
	base := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	if base != "" {
		return base + "/api/sf6/character/" + toolName + ".png"
	}
	return "https://www.streetfighter.com/6/buckler/assets/images/material/character/profile_" + toolName + ".png"
}

func imageExists(ctx context.Context, url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s err=%v\n", url, err)
		}
		return true
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.streetfighter.com/6/buckler/ja-jp/")
	resp, err := client.Do(req)
	if err != nil {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s err=%v\n", url, err)
		}
		return true
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s status=%d (not found)\n", url, resp.StatusCode)
		}
		return false
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s status=%d (ok)\n", url, resp.StatusCode)
		}
		return true
	}
	if resp.StatusCode == http.StatusForbidden {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s status=%d (forbidden)\n", url, resp.StatusCode)
		}
		return true
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists head: url=%s status=%d (fallback ok)\n", url, resp.StatusCode)
		}
		return true
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists get: url=%s err=%v\n", url, err)
		}
		return true
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.streetfighter.com/6/buckler/ja-jp/")
	resp, err = client.Do(req)
	if err != nil {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists get: url=%s err=%v\n", url, err)
		}
		return true
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		if sf6DebugEnabled() {
			fmt.Printf("[sf6][embed] imageExists get: url=%s status=%d (not found)\n", url, resp.StatusCode)
		}
		return false
	}
	if sf6DebugEnabled() {
		fmt.Printf("[sf6][embed] imageExists get: url=%s status=%d\n", url, resp.StatusCode)
	}
	return resp.StatusCode >= 200 && resp.StatusCode < 400 || resp.StatusCode == http.StatusForbidden
}

func sf6DebugEnabled() bool {
	if envBool("SF6_DEBUG") {
		return true
	}
	return envBool("BUCKLER_DEBUG")
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	user := interactionUser(i)
	if user == nil {
		return ""
	}
	return user.ID
}

func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i == nil {
		return nil
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	if i.User != nil {
		return i.User
	}
	return nil
}

func parseUserMention(value string) (string, bool) {
	v := strings.TrimSpace(value)
	if !strings.HasPrefix(v, "<@") || !strings.HasSuffix(v, ">") {
		return "", false
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(v, "<@"), ">")
	inner = strings.TrimPrefix(inner, "!")
	if inner == "" {
		return "", false
	}
	for _, r := range inner {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return inner, true
}

func parseOwnedCustomID(customID, prefix string) (string, bool) {
	if customID == prefix {
		return "", true
	}
	if !strings.HasPrefix(customID, prefix+":") {
		return "", false
	}
	return strings.TrimPrefix(customID, prefix+":"), true
}

func deferEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func deferPublic(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func followupEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func formatJST(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return t.Format("2006-01-02 15:04")
	}
	return t.In(loc).Format("2006-01-02 15:04")
}

func modalValue(components []discordgo.MessageComponent, customID string) string {
	for _, c := range components {
		var row *discordgo.ActionsRow
		switch v := c.(type) {
		case discordgo.ActionsRow:
			row = &v
		case *discordgo.ActionsRow:
			row = v
		default:
			continue
		}
		if row == nil {
			continue
		}
		for _, inner := range row.Components {
			switch input := inner.(type) {
			case discordgo.TextInput:
				if input.CustomID == customID {
					return input.Value
				}
			case *discordgo.TextInput:
				if input != nil && input.CustomID == customID {
					return input.Value
				}
			}
		}
	}
	return ""
}
