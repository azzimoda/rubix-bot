package bot

import (
	"bytes"
	"context"
	"strings"

	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// validMove reports whether the given move name matches one of the buttons
// shown on the board.
func validMove(moveName string) bool {
	for _, row := range moveButtons {
		for _, btn := range row {
			if btn.CallbackData == "move:"+moveName {
				return true
			}
		}
	}
	return false
}

// boardTarget holds the identifiers needed to edit a board message.
type boardTarget struct {
	chatID    int64
	messageID int
	threadID  int
}

// boardTargetFrom extracts the message identifiers from a callback query. The
// boolean is false when the originating message is inaccessible.
func boardTargetFrom(cq *models.CallbackQuery) (boardTarget, bool) {
	if cq.Message.Message == nil {
		return boardTarget{}, false
	}
	m := cq.Message.Message
	return boardTarget{chatID: m.Chat.ID, messageID: m.ID, threadID: m.MessageThreadID}, true
}

// handleCQSession handles session lifecycle callbacks (restart).
func (h *handler) handleCQSession(ctx context.Context, b *bot.Bot, update *models.Update) {
	cq := update.CallbackQuery
	target, ok := boardTargetFrom(cq)
	if !ok {
		answerCallback(ctx, b, cq, "The message is no longer available.")
		return
	}

	chatID, ok := h.resolveActiveChat(ctx, target.chatID)
	if !ok {
		answerCallback(ctx, b, cq, "No active game. Send /start to begin.")
		return
	}

	switch strings.TrimPrefix(cq.Data, "session:") {
	case "restart":
		session, err := h.session.Start(ctx, chatID)
		if err != nil {
			log.Error().Err(err).Msg("restart: failed to start new session")
			answerCallback(ctx, b, cq, "Failed to restart, try /start.")
			return
		}
		if err := h.editBoard(ctx, b, target, session); err != nil {
			log.Error().Err(err).Msg("restart: failed to update message")
		}
		answerCallback(ctx, b, cq, "New game started.")
	default:
		answerCallback(ctx, b, cq, "Unknown action.")
	}
}

// handleCQMove applies a move to the active session and redraws the board.
func (h *handler) handleCQMove(ctx context.Context, b *bot.Bot, update *models.Update) {
	cq := update.CallbackQuery
	target, ok := boardTargetFrom(cq)
	if !ok {
		answerCallback(ctx, b, cq, "The message is no longer available.")
		return
	}

	chatID, ok := h.resolveActiveChat(ctx, target.chatID)
	if !ok {
		answerCallback(ctx, b, cq, "No active game. Send /start to begin.")
		return
	}

	moveName := strings.TrimPrefix(cq.Data, "move:")
	if !validMove(moveName) {
		answerCallback(ctx, b, cq, "Unknown move.")
		return
	}
	session, err := h.session.ApplyMove(ctx, chatID, moveName)
	if err != nil {
		log.Error().Err(err).Msg("move: failed to apply move")
		answerCallback(ctx, b, cq, "Failed to apply move.")
		return
	}

	if err := h.editBoard(ctx, b, target, session); err != nil {
		log.Error().Err(err).Msg("move: failed to update message")
	}

	if session.Solved {
		answerCallback(ctx, b, cq, "🎉 Solved! Great job!")
		return
	}
	answerCallback(ctx, b, cq, "Move applied: "+moveName)
}

// resolveActiveChat returns the internal chat ID for a Telegram chat. The
// second return is false when the chat is not registered.
func (h *handler) resolveActiveChat(ctx context.Context, tgChatID int64) (uint, bool) {
	chat, err := h.chat.GetByTgChatID(ctx, tgChatID)
	if err != nil {
		log.Error().Err(err).Int64("tgChatID", tgChatID).Msg("resolve chat failed")
		return 0, false
	}
	return chat.ID, true
}

// editBoard re-renders the board image, caption and keyboard into the given
// message, replacing the previous state.
func (h *handler) editBoard(ctx context.Context, b *bot.Bot, target boardTarget, session *model.Session) error {
	img, err := BoardImage(session)
	if err != nil {
		return err
	}

	media := &models.InputMediaPhoto{
		Media:           "attach://board.png",
		MediaAttachment: bytes.NewReader(img),
	}
	if err := editMessageMediaRetry(ctx, b, &bot.EditMessageMediaParams{
		ChatID:      target.chatID,
		MessageID:   target.messageID,
		Media:       media,
		ReplyMarkup: boardKeyboard(session.Solved),
	}); err != nil {
		return err
	}

	return editMessageCaptionRetry(ctx, b, &bot.EditMessageCaptionParams{
		ChatID:      target.chatID,
		MessageID:   target.messageID,
		Caption:     boardCaption(session),
		ReplyMarkup: boardKeyboard(session.Solved),
	})
}

func answerCallback(ctx context.Context, b *bot.Bot, cq *models.CallbackQuery, text string) {
	if err := answerCallbackRetry(ctx, b, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cq.ID,
		Text:            text,
	}); err != nil {
		log.Error().Err(err).Msg("failed to answer callback query")
	}
}
