package bot

import (
	"context"
	"errors"
	"fmt"

	"github.com/azzimoda/rubix-bot/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

type handler struct {
	chat    *service.ChatService
	session *service.SessionService
}

func (h *handler) registerHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, h.handleCmdStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommandStartOnly, h.handleCmdHelp)
	b.RegisterHandler(bot.HandlerTypeMessageText, "stats", bot.MatchTypeCommandStartOnly, h.handleCmdStats)
	b.RegisterHandler(bot.HandlerTypeMessageText, "stop", bot.MatchTypeCommandStartOnly, h.handleCmdStop)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "move:", bot.MatchTypePrefix, h.handleCQMove)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "session:", bot.MatchTypePrefix, h.handleCQSession)
}

func (*handler) handleDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Debug().Msg("Unhandled update")
}

func (h *handler) handleCmdStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	chat, err := h.chat.GetByTgChatID(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Error().Err(err).Int64("tgChatID", update.Message.Chat.ID).Msg("start: failed to resolve chat")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
			Text: "Something went wrong, try /start again.",
		})
		return
	}

	session, err := h.session.Start(ctx, chat.ID)
	if err != nil {
		log.Error().Err(err).Msg("start: failed to start session")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
			Text: "Failed to start a new game, please try again later.",
		})
		return
	}

	h.sendBoard(ctx, b, update.Message.Chat.ID, update.Message.MessageThreadID, session)
	log.Info().Msg("Handled command start")
}

func (*handler) handleCmdHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	const helpMsg = "Solve the Rubik's cube by tapping the move buttons.\n" +
		"/start — begin a new game\n" +
		"/stats — show your finished-game statistics\n" +
		"/stop  — end your current game"

	SendMessageRetry(ctx, b, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
		Text: helpMsg,
	})
	log.Info().Msg("Handled command help")
}

// handleCmdStats replies with summary statistics for the chat's finished games.
func (h *handler) handleCmdStats(ctx context.Context, b *bot.Bot, update *models.Update) {
	chat, err := h.chat.GetByTgChatID(ctx, update.Message.Chat.ID)
	if err != nil {
		if errors.Is(err, service.ErrChatNotFound) {
			SendMessageRetry(ctx, b, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
				Text: "No finished games yet. Send /start to play.",
			})
			return
		}
		log.Error().Err(err).Int64("tgChatID", update.Message.Chat.ID).Msg("stats: failed to resolve chat")
		return
	}

	stats, err := h.session.Stats(ctx, chat.ID)
	if err != nil {
		log.Error().Err(err).Msg("stats: failed to compute statistics")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
			Text: "Failed to load your statistics, please try again later.",
		})
		return
	}

	if stats.Total == 0 {
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
			Text: "No finished games yet. Send /start to play.",
		})
		return
	}

	text := fmt.Sprintf("📊 Finished games: %d\n", stats.Total)
	if stats.Solved > 0 {
		text += fmt.Sprintf("Solved: %d (%.0f%%) | Abandoned: %d\n",
			stats.Solved, stats.SolveRate()*100, stats.Abandoned)
		text += fmt.Sprintf("Avg moves to solve: %d\n", stats.AvgMovesSolved)
		text += fmt.Sprintf("Best: %d moves", stats.BestMovesSolved)
	} else {
		text += fmt.Sprintf("Solved: 0 | Abandoned: %d\nNo solved games yet.", stats.Abandoned)
	}

	SendMessageRetry(ctx, b, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
		Text: text,
	})
	log.Info().Msg("Handled command stats")
}

func (h *handler) handleCmdStop(ctx context.Context, b *bot.Bot, update *models.Update) {
	chat, err := h.chat.GetByTgChatID(ctx, update.Message.Chat.ID)
	if err != nil {
		if errors.Is(err, service.ErrChatNotFound) {
			SendMessageRetry(ctx, b, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
				Text: "You have no active game.",
			})
			return
		}
		log.Error().Err(err).Msg("stop: failed to get chat")
		return
	}

	if err := h.session.End(ctx, chat.ID); err != nil {
		if errors.Is(err, service.ErrNoSession) {
			SendMessageRetry(ctx, b, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
				Text: "You have no active game.",
			})
			return
		}
		log.Error().Err(err).Msg("stop: failed to end session")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
			Text: "Failed to end your game, please try again later.",
		})
		return
	}

	SendMessageRetry(ctx, b, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
		Text: "Your current game has been ended. Finished games are kept. Send /start to play again.",
	})
	log.Info().Msg("Handled command stop")
}
