package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ErrMissingToken is returned when BOT_TOKEN is not set.
var ErrMissingToken = errors.New("BOT_TOKEN is not set")

// Config holds the runtime configuration for the bot, loaded from environment
// variables.
type Config struct {
	// BotToken is the Telegram bot API token.
	BotToken string
	// LogLevel is the zerolog level (e.g. "trace", "info", "warn").
	LogLevel string
	// DBPath is the path to the SQLite database file.
	DBPath string
	// HistoryLimitPerChat is the maximum number of ended sessions kept per chat.
	HistoryLimitPerChat int
	// ShutdownTimeout is how long to wait for in-flight work on shutdown.
	ShutdownTimeout time.Duration
	// SOCKS5Proxy is an optional SOCKS5 proxy address ("host:port" or
	// "socks5://host:port"). Empty means a direct connection.
	SOCKS5Proxy string
	// SOCKS5User is the optional username for an authenticated SOCKS5 proxy.
	SOCKS5User string
	// SOCKS5Pass is the optional password for an authenticated SOCKS5 proxy.
	SOCKS5Pass string
}

// Load reads configuration from the environment. Values not present fall back
// to sane defaults. Returns a non-nil error when required values are missing.
func Load() (*Config, error) {
	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		DBPath:      os.Getenv("DB_PATH"),
		SOCKS5Proxy: os.Getenv("SOCKS5_PROXY"),
		SOCKS5User:  os.Getenv("SOCKS5_USER"),
		SOCKS5Pass:  os.Getenv("SOCKS5_PASS"),
	}

	if cfg.BotToken == "" {
		return nil, ErrMissingToken
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "rubix.db"
	}

	limit := os.Getenv("HISTORY_LIMIT_PER_CHAT")
	if limit == "" {
		cfg.HistoryLimitPerChat = 500
	} else {
		n, err := strconv.Atoi(limit)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("HISTORY_LIMIT_PER_CHAT must be a non-negative integer, got %q", limit)
		}
		cfg.HistoryLimitPerChat = n
	}

	timeout := os.Getenv("SHUTDOWN_TIMEOUT")
	if timeout == "" {
		cfg.ShutdownTimeout = 5 * time.Second
	} else {
		d, err := time.ParseDuration(timeout)
		if err != nil || d < 0 {
			return nil, fmt.Errorf("SHUTDOWN_TIMEOUT must be a duration (e.g. 5s), got %q", timeout)
		}
		cfg.ShutdownTimeout = d
	}

	if cfg.SOCKS5Proxy != "" {
		addr, err := proxyAddress(cfg.SOCKS5Proxy)
		if err != nil {
			return nil, err
		}
		cfg.SOCKS5Proxy = addr
	}

	return cfg, nil
}

// proxyAddress accepts "host:port" or "socks5://host:port" and normalizes it to
// "host:port". It rejects other schemes and empty addresses.
func proxyAddress(raw string) (string, error) {
	const scheme = "socks5://"
	if strings.HasPrefix(raw, scheme) {
		raw = strings.TrimPrefix(raw, scheme)
	}
	if raw == "" {
		return "", errors.New("SOCKS5_PROXY must not be empty")
	}
	if strings.Contains(raw, "://") {
		return "", fmt.Errorf("SOCKS5_PROXY must be socks5 (got %q); only the socks5:// scheme is supported", raw)
	}
	return raw, nil
}
