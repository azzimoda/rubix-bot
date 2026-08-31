package database

import (
	"fmt"

	"github.com/azzimoda/rubix-bot/internal/config"
	"github.com/azzimoda/rubix-bot/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens the SQLite database, applies the schema and returns the GORM
// handle. It uses an embedded driver, so no external database server is
// required for the MVP; the repository/service layers keep the code portable
// to PostgreSQL later.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.AutoMigrate(&model.Chat{}, &model.Session{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
