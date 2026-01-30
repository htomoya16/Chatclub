package buckler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CardResponse は Buckler のカードAPIレスポンス（必要項目のみ）。
type CardResponse struct {
	SID                      int64  `json:"sid"`
	FighterName              string `json:"fighter_name"`
	FavoriteCharacterTool    string `json:"favorite_character_tool_name"`
	PlatformToolName         string `json:"platform_tool_name"`
	HomeName                 string `json:"home_name"`
	TitleFileName            string `json:"title_file_name"`
	CircleName               string `json:"circle_name"`
	IconFileName             string `json:"icon_file_name"`
}

// FetchCard は Buckler のカード情報を取得する。
func (c *Client) FetchCard(ctx context.Context, sid string) (CardResponse, error) {
	if sid == "" {
		return CardResponse{}, fmt.Errorf("sid is required")
	}
	base := strings.TrimRight(c.cfg.BucklerBaseURL, "/")
	url := fmt.Sprintf("%s/api/%s/card/%s", base, c.cfg.Lang, sid)
	resp, body, err := c.getReturn(ctx, url)
	if err != nil {
		return CardResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return CardResponse{}, fmt.Errorf("card api status=%d", resp.StatusCode)
	}
	var card CardResponse
	if err := json.Unmarshal(body, &card); err != nil {
		return CardResponse{}, err
	}
	return card, nil
}
