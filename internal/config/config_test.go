package config

import (
	"errors"
	"testing"
)

func TestLoadMissingToken(t *testing.T) {
	t.Setenv("BOT_TOKEN", "")
	_, err := Load()
	if !errors.Is(err, ErrMissingToken) {
		t.Fatalf("expected ErrMissingToken, got %v", err)
	}
}

func TestHistoryLimitDefault(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("HISTORY_LIMIT_PER_CHAT", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HistoryLimitPerChat != 500 {
		t.Fatalf("expected default history limit 500, got %d", cfg.HistoryLimitPerChat)
	}
}

func TestHistoryLimitParsed(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("HISTORY_LIMIT_PER_CHAT", "1200")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HistoryLimitPerChat != 1200 {
		t.Fatalf("expected history limit 1200, got %d", cfg.HistoryLimitPerChat)
	}
}
