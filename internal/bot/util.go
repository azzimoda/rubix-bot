package bot

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const defaultAttempts = 5

// retryPolicy builds the shared retry configuration used for all outbound
// Telegram calls. It retries transient failures with exponential backoff and a
// hard cap, and stops immediately once the context is cancelled (e.g. during
// shutdown) instead of spinning on a dying context.
func retryPolicy(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Attempts(defaultAttempts),
		retry.Delay(200 * time.Millisecond),
		retry.MaxDelay(2 * time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(func(err error) bool {
			return !errors.Is(err, context.Canceled) && ctx.Err() == nil
		}),
	}
}

// doWithRetry runs op through the shared retry policy and returns the outcome
// after the attempts are exhausted.
func doWithRetry(ctx context.Context, op func() error) error {
	return retry.New(retryPolicy(ctx)...).Do(op)
}

func SendMessageRetry(ctx context.Context, b *bot.Bot, params *bot.SendMessageParams) (msg *models.Message, err error) {
	err = doWithRetry(ctx, func() (err error) {
		msg, err = b.SendMessage(ctx, params)
		return
	})
	return
}

func answerCallbackRetry(ctx context.Context, b *bot.Bot, params *bot.AnswerCallbackQueryParams) error {
	return doWithRetry(ctx, func() error {
		_, err := b.AnswerCallbackQuery(ctx, params)
		return err
	})
}

func editMessageMediaRetry(ctx context.Context, b *bot.Bot, params *bot.EditMessageMediaParams) error {
	return doWithRetry(ctx, func() error {
		_, err := b.EditMessageMedia(ctx, params)
		return err
	})
}

func editMessageCaptionRetry(ctx context.Context, b *bot.Bot, params *bot.EditMessageCaptionParams) error {
	return doWithRetry(ctx, func() error {
		_, err := b.EditMessageCaption(ctx, params)
		return err
	})
}
