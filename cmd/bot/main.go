package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/azzimoda/rubix-bot/internal/bot"
	"github.com/azzimoda/rubix-bot/internal/config"
	"github.com/azzimoda/rubix-bot/internal/database"
	"github.com/azzimoda/rubix-bot/internal/repository"
	"github.com/azzimoda/rubix-bot/internal/service"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: Dockerfile & compose.yaml

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Failed to load dotenv file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Missing required configuration")
	}

	setupLogging(cfg.LogLevel)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	chatService := service.NewChatService(repository.NewChatRepository(db))
	sessionService := service.NewSessionService(repository.NewSessionRepository(db), cfg.HistoryLimitPerChat)

	b, err := bot.New(cfg.BotToken, bot.Deps{
		Chat:    chatService,
		Session: sessionService,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bot!")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Run the bot's long-polling loop in the background; it blocks until the
	// context it is given is cancelled, then returns and signals done.
	botDone := make(chan struct{})
	log.Info().Msg("Starting listening...")
	go func() {
		b.Start(ctx)
		close(botDone)
	}()

	<-ctx.Done()
	log.Info().Msg("Shutting down...")

	// Give in-flight requests a bounded grace period, then close the DB.
	if cfg.ShutdownTimeout > 0 {
		select {
		case <-time.After(cfg.ShutdownTimeout):
			log.Warn().Dur("timeout", cfg.ShutdownTimeout).Msg("Shutdown timed out")
		case <-botDone:
		}
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Info().Msg("Done")
}

// setupLogging configures the global zerolog level from the given string.
func setupLogging(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Warn().Str("level", level).Msg("Invalid log level, using info")
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}
