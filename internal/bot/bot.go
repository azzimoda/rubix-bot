package bot

import (
	"fmt"

	"github.com/azzimoda/rubix-bot/internal/service"
	"github.com/go-telegram/bot"
)

// Deps bundles the dependencies required by the bot's handlers.
type Deps struct {
	Chat    *service.ChatService
	Session *service.SessionService
}

// New builds and configures the bot with the given dependencies. The token is
// used for a direct connection; the go-tg-proxy layer can be dropped in
// here later without changing the handlers.
func New(token string, deps Deps) (*bot.Bot, error) {
	h := handler{
		chat:    deps.Chat,
		session: deps.Session,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(h.handleDefault),
		bot.WithMiddlewares(h.ensureUser),
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	h.registerHandlers(b)

	return b, nil
}
