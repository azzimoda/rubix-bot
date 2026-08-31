package bot

import (
	"context"
	"errors"

	"github.com/azzimoda/rubix-bot/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// ensureUser guarantees a Chat record exists for the sender, creating one on
// first contact, before delegating to the wrapped handler.
func (h *handler) ensureUser(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var tgChatID int64
		if m := update.Message; m != nil {
			tgChatID = m.Chat.ID
		} else if cq := update.CallbackQuery; cq != nil && cq.Message.Message != nil {
			tgChatID = cq.Message.Message.Chat.ID
		} else {
			next(ctx, b, update)
			return
		}

		_, err := h.chat.GetByTgChatID(ctx, tgChatID)
		if err == nil {
			next(ctx, b, update)
			return
		}
		if !errors.Is(err, service.ErrChatNotFound) {
			log.Error().Err(err).Int64("tgChatID", tgChatID).Msg("failed to check chat")
			SendMessageRetry(ctx, b, &bot.SendMessageParams{
				ChatID: tgChatID,
				Text:   "Something went wrong, please try again.",
			})
			return
		}

		if _, err := h.chat.Create(ctx, tgChatID); err != nil {
			log.Error().Err(err).Int64("tgChatID", tgChatID).Msg("failed to create chat")
			SendMessageRetry(ctx, b, &bot.SendMessageParams{
				ChatID: tgChatID,
				Text:   "Something went wrong, please try again.",
			})
			return
		}
		log.Info().Int64("tgChatID", tgChatID).Msg("Chat registered")

		next(ctx, b, update)
	}
}
