package api

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var sf6ToolNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

type SF6AssetHandler struct {
	client *http.Client
}

func NewSF6AssetHandler() *SF6AssetHandler {
	return &SF6AssetHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *SF6AssetHandler) CharacterImage(c echo.Context) error {
	tool := strings.TrimSpace(c.Param("tool"))
	if tool == "" {
		return c.NoContent(http.StatusNotFound)
	}
	tool = strings.ToLower(strings.TrimSuffix(tool, ".png"))
	if !sf6ToolNamePattern.MatchString(tool) {
		return c.NoContent(http.StatusNotFound)
	}

	url := "https://www.streetfighter.com/6/buckler/assets/images/material/character/profile_" + tool + ".png"
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, url, nil)
	if err != nil {
		return c.NoContent(http.StatusBadGateway)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.streetfighter.com/6/buckler/ja-jp/")
	resp, err := h.client.Do(req)
	if err != nil {
		return c.NoContent(http.StatusBadGateway)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return c.NoContent(http.StatusNotFound)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return c.NoContent(http.StatusBadGateway)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err = io.Copy(c.Response().Writer, resp.Body)
	return err
}
