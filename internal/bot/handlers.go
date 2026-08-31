package bot

import (
	"context"
	"errors"

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
		"/stop  — delete your data"

	SendMessageRetry(ctx, b, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID, MessageThreadID: update.Message.MessageThreadID,
		Text: helpMsg,
	})
	log.Info().Msg("Handled command help")
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
