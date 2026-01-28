package buckler

import (
	"errors"
	"os"
	"time"
)

// Config holds runtime settings for Buckler access and login.
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
	BuildIDTTL      time.Duration
	UserAgent       string
	CookieEncKeyRaw string
}

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
		BuildIDTTL:     24 * time.Hour,
		UserAgent:      envOrDefault("BUCKLER_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		CookieEncKeyRaw: os.Getenv("BUCKLER_COOKIE_ENC_KEY"),
	}

	if cfg.Email == "" || cfg.Password == "" {
		return cfg, errors.New("CAPCOM_EMAIL/CAPCOM_PASSWORD required")
	}

	return cfg, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
