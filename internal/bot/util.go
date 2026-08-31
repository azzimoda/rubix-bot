package bot

import (
	"context"

	"github.com/avast/retry-go/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var attempts uint = 5

func Attempts() uint       { return attempts }
func SetAttempts(val uint) { attempts = val }

func SendMessageRetry(ctx context.Context, b *bot.Bot, params *bot.SendMessageParams) (msg *models.Message, err error) {
	err = retry.New(retry.Attempts(Attempts())).Do(func() (err error) {
		msg, err = b.SendMessage(ctx, params)
		return
	})
	return
}
