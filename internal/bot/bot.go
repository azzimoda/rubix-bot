package bot

import (
	"fmt"
	"net/http"
	"time"

	"github.com/azzimoda/rubix-bot/internal/service"
	"github.com/go-telegram/bot"
	"golang.org/x/net/proxy"
)

// Deps bundles the dependencies required by the bot's handlers.
type Deps struct {
	Chat    *service.ChatService
	Session *service.SessionService
	// Socks5Addr is an optional SOCKS5 proxy address ("host:port"). Empty
	// means a direct connection to Telegram.
	Socks5Addr string
	// Socks5User and Socks5Pass are optional credentials for an authenticated
	// SOCKS5 proxy.
	Socks5User string
	Socks5Pass string
}

// New builds and configures the bot with the given dependencies. When a SOCKS5
// proxy is configured the bot's HTTP client is routed through it; otherwise a
// direct connection is used.
func New(token string, deps Deps) (*bot.Bot, error) {
	h := handler{
		chat:    deps.Chat,
		session: deps.Session,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(h.handleDefault),
		bot.WithMiddlewares(h.ensureUser),
	}

	if deps.Socks5Addr != "" {
		client, err := socks5Client(deps.Socks5Addr, deps.Socks5User, deps.Socks5Pass)
		if err != nil {
			return nil, fmt.Errorf("failed to build SOCKS5 client: %w", err)
		}
		opts = append(opts, bot.WithHTTPClient(30*time.Second, client))
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	h.registerHandlers(b)

	return b, nil
}

// socks5Client builds an http.Client whose transport dials through the given
// SOCKS5 proxy. Auth is applied only when a username is provided.
func socks5Client(addr, user, pass string) (*http.Client, error) {
	var auth *proxy.Auth
	if user != "" {
		auth = &proxy.Auth{User: user, Password: pass}
	}
	dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	dialContext, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("SOCKS5 dialer does not support context dialing")
	}

	transport := &http.Transport{
		DialContext:         dialContext.DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}
