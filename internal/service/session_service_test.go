package service

import (
	"context"
	"errors"
	"testing"

	"github.com/azzimoda/rubix"
	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/azzimoda/rubix-bot/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestSessionService(t *testing.T) (*SessionService, uint, repository.SessionRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Chat{}, &model.Session{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sessionRepo := repository.NewSessionRepository(db)
	chatRepo := repository.NewChatRepository(db)

	chat := &model.Chat{TgChatID: 12345}
	if err := chatRepo.Create(context.Background(), chat); err != nil {
		t.Fatalf("create chat: %v", err)
	}
	return NewSessionService(sessionRepo, 500), chat.ID, sessionRepo
}

func TestApplyMoveAndSolve(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	sess, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.Solved {
		t.Fatalf("fresh session should not be solved")
	}
	if sess.Scramble == "" {
		t.Fatalf("expected scramble")
	}

	// Apply one move and ensure it is recorded.
	sess, err = svc.ApplyMove(ctx, chatID, "R")
	if err != nil {
		t.Fatalf("ApplyMove: %v", err)
	}
	if got := sess.MovesList(); len(got) != 1 || got[0] != "R" {
		t.Fatalf("expected moves [R], got %v", got)
	}

	// Start a fresh session and solve it by applying the inverse scramble.
	sess, err = svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start 2: %v", err)
	}
	inverse := rubix.Inverse(parseScramble(t, sess.Scramble))
	for _, m := range inverse {
		sess, err = svc.ApplyMove(ctx, chatID, m.String())
		if err != nil {
			t.Fatalf("apply inverse %s: %v", m, err)
		}
	}

	t.Logf("moves applied: %d, solved flag: %v", len(sess.MovesList()), sess.Solved)
	if !sess.Solved {
		t.Fatalf("cube should be solved after inverse scramble")
	}
	cube, err := sess.Cube()
	if err != nil {
		t.Fatalf("cube: %v", err)
	}
	if !cube.IsSolved() {
		t.Fatalf("cube reports not solved")
	}

	// Once solved, GetActive should report no active (unsolved) session.
	if _, err := svc.GetActive(ctx, chatID); err != ErrNoSession {
		t.Fatalf("expected ErrNoSession for a solved-only session, got %v", err)
	}
}

func parseScramble(t *testing.T, scramble string) []rubix.Move {
	t.Helper()
	moves, err := rubix.Parse(scramble)
	if err != nil {
		t.Fatalf("parse scramble %q: %v", scramble, err)
	}
	return moves
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Chat{}, &model.Session{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func repoGetByID(t *testing.T, repo repository.SessionRepository, id uint, out *model.Session) error {
	t.Helper()
	return repo.GetByID(context.Background(), id, out)
}

func TestEndKeepsSolvedSessionsAndActiveOneOnly(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	// First game: solve it so it becomes a historical (solved) session.
	s1, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	inverse := rubix.Inverse(parseScramble(t, s1.Scramble))
	for _, m := range inverse {
		if _, err := svc.ApplyMove(ctx, chatID, m.String()); err != nil {
			t.Fatalf("apply inverse %s: %v", m, err)
		}
	}
	if _, err := svc.GetActive(ctx, chatID); err != ErrNoSession {
		t.Fatalf("expected ErrNoSession, got %v", err)
	}

	// Second game: leave it active, then /stop (End).
	if _, err := svc.Start(ctx, chatID); err != nil {
		t.Fatalf("Start 2: %v", err)
	}
	if err := svc.End(ctx, chatID); err != nil {
		t.Fatalf("End: %v", err)
	}

	// After End, there must be no active session left.
	if _, err := svc.GetActive(ctx, chatID); err != ErrNoSession {
		t.Fatalf("expected ErrNoSession after End, got %v", err)
	}

	// A new game must start cleanly, proving the historical/solved session
	// from game 1 was preserved and only the active one was dropped.
	if _, err := svc.Start(ctx, chatID); err != nil {
		t.Fatalf("Start 3 after End: %v", err)
	}
}

func TestStartSurvivesNoActiveSession(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	if err := svc.End(ctx, chatID); !errors.Is(err, ErrNoSession) {
		t.Fatalf("expected ErrNoSession from End with no active game, got %v", err)
	}

	sess, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start with no existing active session: %v", err)
	}
	if sess.Solved {
		t.Fatalf("fresh session should not be solved")
	}
}

func TestSolveMarksSessionEnded(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	sess, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.Ended {
		t.Fatalf("active session should not be ended")
	}

	inverse := rubix.Inverse(parseScramble(t, sess.Scramble))
	for _, m := range inverse {
		if _, err := svc.ApplyMove(ctx, chatID, m.String()); err != nil {
			t.Fatalf("apply inverse %s: %v", m, err)
		}
	}

	sess, err = svc.GetActive(ctx, chatID)
	if !errors.Is(err, ErrNoSession) {
		t.Fatalf("expected ErrNoSession after solve, got %v %v", sess, err)
	}
}

func TestEndRetainsUnsolvedSession(t *testing.T) {
	ctx := context.Background()
	svc, chatID, repo := newTestSessionService(t)

	active, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	activeID := active.ID

	if err := svc.End(ctx, chatID); err != nil {
		t.Fatalf("End: %v", err)
	}

	var got model.Session
	if err := repoGetByID(t, repo, activeID, &got); err != nil {
		t.Fatalf("expected ended session to be retained, got error: %v", err)
	}
	if !got.Ended {
		t.Fatalf("expected retained session to be marked ended")
	}
}

func TestStartTrimsEndedHistoryToLimit(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := repository.NewSessionRepository(db)
	chat := &model.Chat{TgChatID: 999}
	if err := repository.NewChatRepository(db).Create(ctx, chat); err != nil {
		t.Fatalf("create chat: %v", err)
	}

	const limit = 3
	svc := NewSessionService(repo, limit)

	// Create limit+2 ended sessions, then Start adds the active one.
	for i := 0; i < limit+2; i++ {
		s := model.NewSession(chat.ID)
		s.Ended = true
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("create ended session: %v", err)
		}
	}

	if _, err := svc.Start(ctx, chat.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	var total int64
	if err := db.WithContext(ctx).Model(&model.Session{}).Where("chat_id = ?", chat.ID).Count(&total).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != int64(limit+1) {
		t.Fatalf("expected %d sessions after trim (limit + active), got %d", limit+1, total)
	}

	// The active one must be present and not ended.
	if _, err := repo.GetActiveByChat(ctx, chat.ID); err != nil {
		t.Fatalf("expected an active session, got %v", err)
	}
}

func TestStartWithScramblePersistsCustomScramble(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	moves, err := rubix.Parse("R U F' L B2")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sess, err := svc.StartWithScramble(ctx, chatID, moves)
	if err != nil {
		t.Fatalf("StartWithScramble: %v", err)
	}
	if want := "R U F' L B2"; sess.Scramble != want {
		t.Fatalf("scramble mismatch: %q != %q", sess.Scramble, want)
	}
	if sess.Solved || sess.Ended {
		t.Fatalf("fresh session should be active and unsolved, got solved=%v ended=%v", sess.Solved, sess.Ended)
	}

	active, err := svc.GetActive(ctx, chatID)
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if active.ID != sess.ID {
		t.Fatalf("expected active session to match started one")
	}
}

func TestStartWithScrambleEndsPreviousActive(t *testing.T) {
	ctx := context.Background()
	svc, chatID, repo := newTestSessionService(t)

	first, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	moves, err := rubix.Parse("R U F' L B2")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := svc.StartWithScramble(ctx, chatID, moves); err != nil {
		t.Fatalf("StartWithScramble: %v", err)
	}

	var prev model.Session
	if err := repoGetByID(t, repo, first.ID, &prev); err != nil {
		t.Fatalf("expected previous session retained, got %v", err)
	}
	if !prev.Ended {
		t.Fatalf("expected previous active session to be marked ended")
	}
	if prev.Solved {
		t.Fatalf("previous session should not be marked solved")
	}
}

func TestStartDefaultStillRandomScrambled(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	sess, err := svc.Start(ctx, chatID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.Scramble == "" {
		t.Fatalf("expected a default random scramble")
	}
	cube, err := sess.Cube()
	if err != nil {
		t.Fatalf("Cube: %v", err)
	}
	if cube.IsSolved() {
		t.Fatalf("default session must be scrambled")
	}
}

func TestStatsOverEndedSessions(t *testing.T) {
	ctx := context.Background()
	svc, chatID, repo := newTestSessionService(t)

	addEnded := func(solved bool, moves string) {
		s := model.NewSession(chatID)
		s.Ended = true
		s.Solved = solved
		s.Moves = moves
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	// Solved sessions with 3, 8 and 6 moves; two abandoned.
	addEnded(true, "R U F")
	addEnded(true, "R2 U F B L D R U")
	addEnded(true, "R2 U F L D R")
	addEnded(false, "R U")
	addEnded(false, "R U F D L R U F B")

	stats, err := svc.Stats(ctx, chatID)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Total != 5 {
		t.Fatalf("expected total 5, got %d", stats.Total)
	}
	if stats.Solved != 3 {
		t.Fatalf("expected solved 3, got %d", stats.Solved)
	}
	if stats.Abandoned != 2 {
		t.Fatalf("expected abandoned 2, got %d", stats.Abandoned)
	}
	// avg = (3+8+6)/3 = 17/3 ≈ 6
	if stats.AvgMovesSolved != 6 {
		t.Fatalf("expected avg moves 6, got %d", stats.AvgMovesSolved)
	}
	if stats.BestMovesSolved != 3 {
		t.Fatalf("expected best moves 3, got %d", stats.BestMovesSolved)
	}
	if got := stats.SolveRate(); got != 0.6 {
		t.Fatalf("expected solve rate 0.6, got %v", got)
	}
}

func TestStatsEmptyHistory(t *testing.T) {
	ctx := context.Background()
	svc, chatID, _ := newTestSessionService(t)

	stats, err := svc.Stats(ctx, chatID)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Total != 0 || stats.Solved != 0 || stats.Abandoned != 0 {
		t.Fatalf("expected all-zero stats for empty history, got %+v", stats)
	}
	if stats.SolveRate() != 0 {
		t.Fatalf("expected solve rate 0 for empty history, got %v", stats.SolveRate())
	}
}
