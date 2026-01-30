package buckler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// BattlelogResponse は Buckler battlelog JSON の最小構造。
type BattlelogResponse struct {
	PageProps BattlelogPageProps `json:"pageProps"`
}

type BattlelogPageProps struct {
	CurrentPage int           `json:"current_page"`
	TotalPage   int           `json:"total_page"`
	SID         int64         `json:"sid"`
	ReplayList  []ReplayEntry `json:"replay_list"`
}

type ReplayEntry struct {
	ReplayID              string     `json:"replay_id"`
	UploadedAt            int64      `json:"uploaded_at"`
	ReplayBattleType      int         `json:"replay_battle_type"`
	ReplayBattleTypeName  string     `json:"replay_battle_type_name"`
	ReplayBattleSubType   int         `json:"replay_battle_sub_type"`
	ReplayBattleSubTypeName string   `json:"replay_battle_sub_type_name"`
	Player1Info           PlayerInfo `json:"player1_info"`
	Player2Info           PlayerInfo `json:"player2_info"`
}

type PlayerInfo struct {
	Player               Player `json:"player"`
	PlayingCharacterID   int    `json:"playing_character_id"`
	PlayingCharacterName string `json:"playing_character_name"`
	CharacterName        string `json:"character_name"`
	CharacterToolName    string `json:"character_tool_name"`
	RoundResults         []int  `json:"round_results"`
}

type Player struct {
	FighterID       string `json:"fighter_id"`
	ShortID         int64  `json:"short_id"`
	PlatformName    string `json:"platform_name"`
	PlatformToolName string `json:"platform_tool_name"`
}

// RoundWins は round_results の「>0」の個数を勝ちラウンド数として数える。
func RoundWins(results []int) int {
	wins := 0
	for _, v := range results {
		if v > 0 {
			wins++
		}
	}
	return wins
}

// FetchCustomBattlelog は Custom Room の battlelog JSON を取得する。
func (c *Client) FetchCustomBattlelog(ctx context.Context, sid string, page int) (BattlelogResponse, error) {
	var res BattlelogResponse
	if sid == "" {
		return res, errors.New("sid required")
	}
	if page <= 0 {
		page = 1
	}

	if err := c.EnsureLogin(ctx); err != nil {
		return res, err
	}

	buildID, err := c.FetchBuildID(ctx, sid)
	if err != nil {
		return res, err
	}
	url := c.buildBattlelogURL(buildID, sid, page)

	resp, body, err := c.getReturn(ctx, url)
	if err != nil {
		return res, err
	}
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		// rebuild buildId and retry once
		c.cache.Set("")
		buildID, err = c.FetchBuildID(ctx, sid)
		if err != nil {
			return res, err
		}
		url = c.buildBattlelogURL(buildID, sid, page)
		_, body, err = c.getReturn(ctx, url)
		if err != nil {
			return res, err
		}
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return res, fmt.Errorf("decode battlelog: %w", err)
	}
	return res, nil
}

func (c *Client) buildBattlelogURL(buildID, sid string, page int) string {
	base := strings.TrimRight(c.cfg.BucklerBaseURL, "/")
	lang := c.cfg.Lang
	return fmt.Sprintf("%s/_next/data/%s/%s/profile/%s/battlelog/custom.json?sid=%s&page=%s", base, buildID, lang, sid, sid, strconv.Itoa(page))
}
