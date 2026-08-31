package bot

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/azzimoda/rubix-bot/internal/render"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const captionMax = 1024

// moveButtons is the inline keyboard layout of the moves a user can apply.
var moveButtons = [][]models.InlineKeyboardButton{
	{
		{Text: "F", CallbackData: "move:F"},
		{Text: "F'", CallbackData: "move:F'"},
		{Text: "B", CallbackData: "move:B"},
		{Text: "B'", CallbackData: "move:B'"},
	},
	{
		{Text: "U", CallbackData: "move:U"},
		{Text: "U'", CallbackData: "move:U'"},
		{Text: "D", CallbackData: "move:D"},
		{Text: "D'", CallbackData: "move:D'"},
	},
	{
		{Text: "L", CallbackData: "move:L"},
		{Text: "L'", CallbackData: "move:L'"},
		{Text: "R", CallbackData: "move:R"},
		{Text: "R'", CallbackData: "move:R'"},
	},
	{
		{Text: "x", CallbackData: "move:x"},
		{Text: "x'", CallbackData: "move:x'"},
		{Text: "y", CallbackData: "move:y"},
		{Text: "y'", CallbackData: "move:y'"},
		{Text: "z", CallbackData: "move:z"},
		{Text: "z'", CallbackData: "move:z'"},
	},
}

// boardKeyboard builds the inline keyboard shown on the board. When the cube
// is solved only the restart button remains.
func boardKeyboard(solved bool) *models.InlineKeyboardMarkup {
	if solved {
		return &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "🔄 Restart", CallbackData: "session:restart"}},
			},
		}
	}
	rows := make([][]models.InlineKeyboardButton, len(moveButtons)+1)
	copy(rows, moveButtons)
	rows[len(moveButtons)] = []models.InlineKeyboardButton{
		{Text: "🔄 Restart", CallbackData: "session:restart"},
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// BoardImage renders the cube to a PNG suitable for sending to Telegram.
func BoardImage(session *model.Session) ([]byte, error) {
	cube, err := session.Cube()
	if err != nil {
		return nil, err
	}
	return render.Board(cube)
}

// sendBoard sends a brand-new board photo with caption and keyboard.
func (h *handler) sendBoard(ctx context.Context, b *bot.Bot, chatID int64, threadID int, session *model.Session) {
	img, err := BoardImage(session)
	if err != nil {
		log.Error().Err(err).Msg("sendBoard: failed to render board")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: chatID, MessageThreadID: threadID,
			Text: "Failed to build the board, please try again later.",
		})
		return
	}

	photo := &models.InputFileUpload{Filename: "board.png", Data: bytes.NewReader(img)}
	if _, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:          chatID,
		MessageThreadID: threadID,
		Photo:           photo,
		Caption:         boardCaption(session),
		ReplyMarkup:     boardKeyboard(session.Solved),
	}); err != nil {
		log.Error().Err(err).Msg("sendBoard: failed to send photo")
	}
}

func boardCaption(session *model.Session) string {
	var b strings.Builder

	if session.Solved {
		b.WriteString("🎉 Solved! 🎉\n\n")
	} else {
		b.WriteString("🧩 Solve the cube by tapping the move buttons.\n\n")
	}

	fmt.Fprintf(&b, "📋 Scramble: %s\n\n", session.Scramble)

	moves := session.MovesList()
	if len(moves) == 0 {
		b.WriteString("🎯 Moves (0): —")
	} else {
		fmt.Fprintf(&b, "🎯 Moves (%d): %s", len(moves), strings.Join(moves, " "))
	}

	s := b.String()
	if len(s) > captionMax {
		s = s[:captionMax]
	}
	return s
}
