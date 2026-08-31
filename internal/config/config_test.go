package config

import (
	"errors"
	"testing"
	"time"
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

func TestShutdownTimeoutDefault(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ShutdownTimeout != 5*time.Second {
		t.Fatalf("expected default shutdown timeout 5s, got %v", cfg.ShutdownTimeout)
	}
}

func TestShutdownTimeoutParsed(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("SHUTDOWN_TIMEOUT", "2s")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ShutdownTimeout != 2*time.Second {
		t.Fatalf("expected shutdown timeout 2s, got %v", cfg.ShutdownTimeout)
	}
}
