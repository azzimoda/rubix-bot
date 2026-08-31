package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
}

// Load reads configuration from the environment. Values not present fall back
// to sane defaults. Returns a non-nil error when required values are missing.
func Load() (*Config, error) {
	cfg := &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		LogLevel: os.Getenv("LOG_LEVEL"),
		DBPath:   os.Getenv("DB_PATH"),
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

	return cfg, nil
}
