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

// moveButtonsRich is the rich-message button layout of the moves a user can
// apply. Rich message buttons wrap naturally, so the layout is a flat list.
var moveButtonsRich = []models.RichMessageButton{
	richButton("F", "move:F"),
	richButton("F'", "move:F'"),
	richButton("B", "move:B"),
	richButton("B'", "move:B'"),
	richButton("U", "move:U"),
	richButton("U'", "move:U'"),
	richButton("D", "move:D"),
	richButton("D'", "move:D'"),
	richButton("L", "move:L"),
	richButton("L'", "move:L'"),
	richButton("R", "move:R"),
	richButton("R'", "move:R'"),
	richButton("x", "move:x"),
	richButton("x'", "move:x'"),
	richButton("y", "move:y"),
	richButton("y'", "move:y'"),
	richButton("z", "move:z"),
	richButton("z'", "move:z'"),
}

// richButton builds a plain-text rich message button with callback data.
func richButton(text, callbackData string) models.RichMessageButton {
	return models.RichMessageButton{Text: models.RichText{PlainText: text}, CallbackData: callbackData}
}

// restartButton is the button shown to start a fresh game.
func restartButton() models.RichMessageButton {
	return richButton("🔄 Restart", "session:restart")
}

// BoardImage renders the cube to a PNG suitable for sending to Telegram.
func BoardImage(session *model.Session) ([]byte, error) {
	cube, err := session.Cube()
	if err != nil {
		return nil, err
	}
	return render.Board(cube)
}

// boardRichMessage builds the rich message presenting the board: a status
// heading, the rendered cube photo with the scramble as its caption, a
// collapsible move log and the move/restart buttons.
func boardRichMessage(session *model.Session) (*models.InputRichMessage, error) {
	img, err := BoardImage(session)
	if err != nil {
		return nil, err
	}

	status := "🧩 Solve the cube"
	if session.Solved {
		status = "🎉 Solved! 🎉"
	}

	blocks := []models.InputRichBlock{
		{
			Type: models.RichBlockTypeSectionHeading,
			InputRichBlockSectionHeading: &models.InputRichBlockSectionHeading{
				Text: models.RichText{PlainText: status},
				Size: 2,
			},
		},
		{
			Type: models.RichBlockTypePhoto,
			InputRichBlockPhoto: &models.InputRichBlockPhoto{
				Photo: models.InputMediaPhoto{
					Media:           "attach://board.png",
					MediaAttachment: bytes.NewReader(img),
				},
				Caption: &models.RichBlockCaption{Text: models.RichText{PlainText: photoCaption(session)}},
			},
		},
		{
			Type: models.RichBlockTypeDetails,
			InputRichBlockDetails: &models.InputRichBlockDetails{
				Summary: models.RichText{PlainText: fmt.Sprintf("🎯 Moves (%d)", len(session.MovesList()))},
				IsOpen:  !session.Solved,
				Blocks:  []models.InputRichBlock{moveLogBlock(session)},
			},
		},
		{
			Type: models.RichBlockTypeButtons,
			InputRichBlockButtons: &models.InputRichBlockButtons{
				Buttons: boardButtons(session.Solved),
			},
		},
	}

	return &models.InputRichMessage{Blocks: blocks}, nil
}

// boardButtons is the button set for the board. While a game is active it shows
// the move buttons plus restart; once solved only restart remains.
func boardButtons(solved bool) []models.RichMessageButton {
	if solved {
		return []models.RichMessageButton{restartButton()}
	}
	buttons := make([]models.RichMessageButton, 0, len(moveButtonsRich)+1)
	buttons = append(buttons, moveButtonsRich...)
	return append(buttons, restartButton())
}

// photoCaption is the caption shown beneath the rendered cube.
func photoCaption(session *model.Session) string {
	return "📋 Scramble: " + session.Scramble
}

// moveLogBlock renders the player's moves so far, or a hint when none exist.
func moveLogBlock(session *model.Session) models.InputRichBlock {
	text := "No moves applied yet."
	if moves := session.MovesList(); len(moves) > 0 {
		text = strings.Join(moves, " ")
	}
	return models.InputRichBlock{
		Type: models.RichBlockTypePreformatted,
		InputRichBlockPreformatted: &models.InputRichBlockPreformatted{
			Text:     models.RichText{PlainText: text},
			Language: "text",
		},
	}
}

// sendBoard sends a brand-new board as a rich message.
func (h *handler) sendBoard(ctx context.Context, b *bot.Bot, chatID int64, threadID int, session *model.Session) {
	rich, err := boardRichMessage(session)
	if err != nil {
		log.Error().Err(err).Msg("sendBoard: failed to render board")
		SendMessageRetry(ctx, b, &bot.SendMessageParams{
			ChatID: chatID, MessageThreadID: threadID,
			Text: "Failed to build the board, please try again later.",
		})
		return
	}

	if _, err := sendRichMessageRetry(ctx, b, &bot.SendRichMessageParams{
		ChatID:          chatID,
		MessageThreadID: threadID,
		RichMessage:     *rich,
	}); err != nil {
		log.Error().Err(err).Msg("sendBoard: failed to send board")
	}
}
