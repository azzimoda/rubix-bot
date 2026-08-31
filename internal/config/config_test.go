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

func TestProxyAddressNormalizes(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"127.0.0.1:9050", "127.0.0.1:9050", true},
		{"socks5://127.0.0.1:9050", "127.0.0.1:9050", true},
		{"http://host:8080", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, err := proxyAddress(c.in)
		if c.ok {
			if err != nil {
				t.Fatalf("proxyAddress(%q): unexpected error %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("proxyAddress(%q) = %q, want %q", c.in, got, c.want)
			}
		} else if err == nil {
			t.Fatalf("proxyAddress(%q): expected error, got %q", c.in, got)
		}
	}
}

func TestLoadSocks5Config(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("SOCKS5_PROXY", "socks5://127.0.0.1:9050")
	t.Setenv("SOCKS5_USER", "u")
	t.Setenv("SOCKS5_PASS", "p")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SOCKS5Proxy != "127.0.0.1:9050" {
		t.Fatalf("expected normalized proxy addr, got %q", cfg.SOCKS5Proxy)
	}
	if cfg.SOCKS5User != "u" || cfg.SOCKS5Pass != "p" {
		t.Fatalf("expected proxy credentials to be read, got %q/%q", cfg.SOCKS5User, cfg.SOCKS5Pass)
	}
}

func TestLoadSocks5InvalidScheme(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("SOCKS5_PROXY", "https://host:8080")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for unsupported proxy scheme")
	}
}
