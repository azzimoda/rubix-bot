# rubix-bot

Telegram bot for solving a Rubik's cube.

## How it works

- `/start` — begins a new game: a fresh 3x3 cube is scrambled and sent as a
  picture with an inline keyboard.
- Tap the move buttons (F, F', B, B', U, U', D, D', L, L', R, R') to rotate the
  cube. The board is redrawn after every move and the full move log is shown in
  the caption.
- Once the cube is solved, the bot congratulates you.
- `/stop` — ends your current (active) game. Finished games are kept for
  reference; send `/start` to play again.
- `/help` — shows available commands.

The picture shows the cube from two orthographic perspectives side by side with
the unfolded net in the corner. Sessions, the scramble and every applied move
are persisted so a game survives restarts.

Ended games — both solved and abandoned — are kept as history for later
analysis. History per chat is bounded to the newest
`HISTORY_LIMIT_PER_CHAT` ended sessions; the active game is never trimmed.

## Configuration

Environment variables (loaded from `.env` if present):

- `BOT_TOKEN` — required, Telegram bot token.
- `LOG_LEVEL` — optional, zerolog level (default `info`).
- `DB_PATH` — optional, path to the SQLite database (default `rubix.db`).
- `HISTORY_LIMIT_PER_CHAT` — optional, ended sessions kept per chat (default `500`).

## Run

```sh
cp .env.example .env   # set BOT_TOKEN
go run ./cmd/bot
```

## Roadmap

- Multiplayer voting mode: several users solve one cube together by voting for
  the next move.
- Optional SOCKS5 proxy layer (go-tg-proxy) when Telegram is blocked.
- Rich-message presentation once the experimental Bot API is broadly supported.
- `/start [scrumble]` starts a new session with specified scrumble.
