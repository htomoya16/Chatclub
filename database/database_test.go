package database

import (
	"net/url"
	"testing"
)

func TestNormalizeDatabaseURLAddsHerokuDefaults(t *testing.T) {
	dsn, err := normalizeDatabaseURL("postgres://user:pass@example.com:5432/chatclub", "require")
	if err != nil {
		t.Fatalf("normalizeDatabaseURL() error = %v", err)
	}

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	q := parsed.Query()
	if got := q.Get("sslmode"); got != "require" {
		t.Fatalf("sslmode = %q, want require", got)
	}
	if got := q.Get("timezone"); got != "UTC" {
		t.Fatalf("timezone = %q, want UTC", got)
	}
}

func TestNormalizeDatabaseURLKeepsExplicitSSLMode(t *testing.T) {
	dsn, err := normalizeDatabaseURL("postgres://user:pass@example.com:5432/chatclub?sslmode=verify-full", "require")
	if err != nil {
		t.Fatalf("normalizeDatabaseURL() error = %v", err)
	}

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if got := parsed.Query().Get("sslmode"); got != "verify-full" {
		t.Fatalf("sslmode = %q, want verify-full", got)
	}
}
