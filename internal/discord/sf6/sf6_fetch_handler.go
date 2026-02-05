package sf6

import (
	"fmt"
	"os"
	"strings"

	"backend/internal/discord/common"

	"github.com/bwmarrin/discordgo"
)

func (r *Handler) handleSF6Fetch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		common.RespondEphemeral(s, i, "guildのみ対応")
		return
	}
	if r.SF6Service == nil || r.SF6AccountService == nil || r.SF6FriendService == nil {
		common.RespondEphemeral(s, i, "sf6機能が無効です（Buckler設定未完）")
		return
	}

	userID := common.InteractionUserID(i)
	if userID == "" {
		common.RespondEphemeral(s, i, "user_idの取得に失敗")
		return
	}
	if !sf6FetchAllowed(i) {
		common.RespondEphemeral(s, i, "管理者または許可ユーザーのみ実行できます")
		return
	}

	if err := common.DeferEphemeral(s, i); err != nil {
		common.RespondEphemeral(s, i, "受付に失敗しました")
		return
	}

	ctx, cancel := common.CommandContextForInteraction(s, i)
	defer cancel()
	accounts, err := r.SF6AccountService.ListByGuild(ctx, i.GuildID)
	if err != nil {
		common.FollowupEphemeral(s, i, "取得に失敗: "+err.Error())
		return
	}
	friends, err := r.SF6FriendService.ListByGuild(ctx, i.GuildID)
	if err != nil {
		common.FollowupEphemeral(s, i, "取得に失敗: "+err.Error())
		return
	}
	if len(accounts) == 0 && len(friends) == 0 {
		common.FollowupEphemeral(s, i, "対象アカウント/フレンドがありません")
		return
	}

	accountSIDs := make(map[string]struct{}, len(accounts))
	for _, acc := range accounts {
		if acc.FighterID == "" || acc.UserID == "" {
			continue
		}
		accountSIDs[acc.FighterID] = struct{}{}
	}

	totalSaved := 0
	pagesFetched := 0
	accountsFetched := 0
	friendsFetched := 0
	skippedFriends := 0
	fetchErrors := 0

	for _, acc := range accounts {
		if acc.FighterID == "" || acc.UserID == "" {
			continue
		}
		saved, pages, err := r.initialFetch(ctx, i.GuildID, acc.UserID, acc.FighterID)
		if err != nil {
			fetchErrors++
			continue
		}
		accountsFetched++
		totalSaved += saved
		pagesFetched += pages
	}

	seenFriends := make(map[string]struct{}, len(friends))
	for _, friend := range friends {
		if friend.FighterID == "" || friend.UserID == "" {
			continue
		}
		if _, ok := accountSIDs[friend.FighterID]; ok {
			skippedFriends++
			continue
		}
		if _, ok := seenFriends[friend.FighterID]; ok {
			continue
		}
		seenFriends[friend.FighterID] = struct{}{}
		saved, pages, err := r.initialFetch(ctx, i.GuildID, friend.UserID, friend.FighterID)
		if err != nil {
			fetchErrors++
			continue
		}
		friendsFetched++
		totalSaved += saved
		pagesFetched += pages
	}

	msg := fmt.Sprintf("取得完了。accounts=%d friends=%d skipped=%d saved=%d pages=%d errors=%d",
		accountsFetched, friendsFetched, skippedFriends, totalSaved, pagesFetched, fetchErrors)
	common.FollowupEphemeral(s, i, msg)
}

func sf6FetchAllowed(i *discordgo.InteractionCreate) bool {
	userID := common.InteractionUserID(i)
	if userID == "" {
		return false
	}
	if sf6FetchAllowedUser(userID) {
		return true
	}
	if i == nil || i.Member == nil {
		return false
	}
	perms := i.Member.Permissions
	if perms&discordgo.PermissionAdministrator != 0 {
		return true
	}
	if perms&discordgo.PermissionManageGuild != 0 {
		return true
	}
	return false
}

func sf6FetchAllowedUser(userID string) bool {
	raw := strings.TrimSpace(os.Getenv("SF6_FETCH_ALLOWED_USER_IDS"))
	if raw == "" {
		return false
	}
	for _, id := range splitIDList(raw) {
		if id == userID {
			return true
		}
	}
	return false
}

func splitIDList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
