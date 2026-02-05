package buckler

import (
	"errors"
	"os"
	"strings"
	"time"
)

// Config は Buckler 取得・ログインのための設定をまとめた構造体。
type Config struct {
	Email           string
	Password        string
	AuthBaseURL     string
	CidBaseURL      string
	BucklerBaseURL  string
	Lang            string
	ClientID        string
	Connection      string
	Tenant          string
	Protocol        string
	RedirectURI     string
	ResponseType    string
	Scope           string
	UILocales       string
	ShowSignUp      string
	Audience        string
	BucklerClientID string
	BucklerRedirect string
	BucklerScope    string
	BucklerAudience string
	BucklerResponse string
	Debug           bool
	BuildIDTTL      time.Duration
	UserAgent       string
	CookieEncKeyRaw string
}

// LoadConfigFromEnv は環境変数から設定を読み込む。
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Email:          os.Getenv("CAPCOM_EMAIL"),
		Password:       os.Getenv("CAPCOM_PASSWORD"),
		AuthBaseURL:    envOrDefault("CAPCOM_AUTH_BASE_URL", "https://auth.cid.capcom.com"),
		CidBaseURL:     envOrDefault("CAPCOM_CID_BASE_URL", "https://cid.capcom.com"),
		BucklerBaseURL: envOrDefault("BUCKLER_BASE_URL", "https://www.streetfighter.com/6/buckler"),
		Lang:           envOrDefault("BUCKLER_LANG", "ja-jp"),
		ClientID:       envOrDefault("CAPCOM_CLIENT_ID", "mVxOARlAyTcJkcFAb8IZoiKYV8qGAH9a"),
		Connection:     envOrDefault("CAPCOM_CONNECTION", "Username-Password-Authentication"),
		Tenant:         envOrDefault("CAPCOM_TENANT", "capcom"),
		Protocol:       envOrDefault("CAPCOM_PROTOCOL", "oauth2"),
		RedirectURI:    envOrDefault("CAPCOM_REDIRECT_URI", "https://cid.capcom.com/ja/loginCallback"),
		ResponseType:   envOrDefault("CAPCOM_RESPONSE_TYPE", "code"),
		Scope:          envOrDefault("CAPCOM_SCOPE", "openid profile email"),
		UILocales:      envOrDefault("CAPCOM_UI_LOCALES", "ja"),
		ShowSignUp:     envOrDefault("CAPCOM_SHOW_SIGN_UP", "0"),
		Audience:       envOrDefault("CAPCOM_AUDIENCE", "urn:rebe:capcom:apis"),
		BucklerClientID: os.Getenv("BUCKLER_CLIENT_ID"),
		BucklerRedirect: envOrDefault("BUCKLER_REDIRECT_URI", "https://www.streetfighter.com/6/buckler/auth/login"),
		BucklerScope:    envOrDefault("BUCKLER_SCOPE", "openid"),
		BucklerAudience: envOrDefault("BUCKLER_AUDIENCE", "urn:rebe:capcom:apis"),
		BucklerResponse: envOrDefault("BUCKLER_RESPONSE_TYPE", "code"),
		Debug:           envBool("BUCKLER_DEBUG", false),
		BuildIDTTL:     24 * time.Hour,
		UserAgent:      envOrDefault("BUCKLER_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		CookieEncKeyRaw: os.Getenv("BUCKLER_COOKIE_ENC_KEY"),
	}

	if cfg.Email == "" || cfg.Password == "" {
		return cfg, errors.New("CAPCOM_EMAIL/CAPCOM_PASSWORD required")
	}

	return cfg, nil
}

// envOrDefault は未設定ならデフォルト値を返す。
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
