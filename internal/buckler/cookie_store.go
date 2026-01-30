package buckler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

// CookieEnvelope はCookieのスナップショットを保存するための構造体。
type CookieEnvelope struct {
	URL     string         `json:"url"`
	Cookies []*http.Cookie `json:"cookies"`
}

// ExportCookies は指定URLのCookieを取り出す。
func (c *Client) ExportCookies(rawURL string) (CookieEnvelope, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return CookieEnvelope{}, err
	}
	cookies := c.client.Jar.Cookies(u)
	return CookieEnvelope{URL: rawURL, Cookies: cookies}, nil
}

// ImportCookies は保存済みCookieをJarへ復元する。
func (c *Client) ImportCookies(env CookieEnvelope) error {
	u, err := url.Parse(env.URL)
	if err != nil {
		return err
	}
	c.client.Jar.SetCookies(u, env.Cookies)
	return nil
}

// EncryptEnvelope はCookieをAES-GCMで暗号化して文字列にする。
func EncryptEnvelope(keyRaw string, env CookieEnvelope) (string, error) {
	key, err := parseCookieKey(keyRaw)
	if err != nil {
		return "", err
	}
	plain, err := json.Marshal(env)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptEnvelope は暗号化されたCookie文字列を復号する。
func DecryptEnvelope(keyRaw, blob string) (CookieEnvelope, error) {
	key, err := parseCookieKey(keyRaw)
	if err != nil {
		return CookieEnvelope{}, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return CookieEnvelope{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return CookieEnvelope{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return CookieEnvelope{}, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return CookieEnvelope{}, errors.New("ciphertext too short")
	}
	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return CookieEnvelope{}, err
	}
	var env CookieEnvelope
	if err := json.Unmarshal(plain, &env); err != nil {
		return CookieEnvelope{}, err
	}
	return env, nil
}

// parseCookieKey は32バイト鍵（raw/base64/hex）を解釈する。
func parseCookieKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, errors.New("cookie encryption key missing")
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, errors.New("cookie encryption key must be 32 bytes (raw/base64/hex)")
}
