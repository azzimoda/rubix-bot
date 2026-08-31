package service

import (
	"context"
	"errors"

	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/azzimoda/rubix-bot/internal/repository"
	"gorm.io/gorm"
)

// ErrNoSession is returned when there is no active session.
var ErrNoSession = errors.New("no active session")

// SessionService manages game sessions and the moves applied to their cubes.
type SessionService struct {
	repo         repository.SessionRepository
	historyLimit int
}

// NewSessionService builds a SessionService. historyLimit is the maximum number
// of ended (solved or abandoned) sessions kept per chat; older ones are trimmed.
func NewSessionService(repo repository.SessionRepository, historyLimit int) *SessionService {
	return &SessionService{repo: repo, historyLimit: historyLimit}
}

// Start creates a new active session for the chat. Any previous active session
// is kept and marked ended; history is then trimmed to the configured limit.
func (s *SessionService) Start(ctx context.Context, chatID uint) (*model.Session, error) {
	if prev, err := s.repo.GetActiveByChat(ctx, chatID); err == nil {
		prev.Ended = true
		if err := s.repo.Update(ctx, prev); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	session := model.NewSession(chatID)
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	if s.historyLimit > 0 {
		if err := s.repo.TrimHistory(ctx, chatID, s.historyLimit); err != nil {
			return nil, err
		}
	}
	return session, nil
}

// GetActive returns the active session for the chat, if any.
func (s *SessionService) GetActive(ctx context.Context, chatID uint) (*model.Session, error) {
	session, err := s.repo.GetActiveByChat(ctx, chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoSession
		}
		return nil, err
	}
	return session, nil
}

// ApplyMove applies a single move name to the cube of the active session,
// records it in the log and marks the session solved when the cube is complete.
func (s *SessionService) ApplyMove(ctx context.Context, chatID uint, moveName string) (*model.Session, error) {
	session, err := s.GetActive(ctx, chatID)
	if err != nil {
		return nil, err
	}

	cube, err := session.Cube()
	if err != nil {
		return nil, err
	}
	cube.Apply(parseMove(moveName))
	session.SetCube(cube)
	session.AppendMove(moveName)
	session.Solved = cube.IsSolved()
	if session.Solved {
		session.Ended = true
	}

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// End closes the active session, marking it ended and keeping it for history.
func (s *SessionService) End(ctx context.Context, chatID uint) error {
	session, err := s.GetActive(ctx, chatID)
	if err != nil {
		return err
	}
	session.Ended = true
	return s.repo.Update(ctx, session)
}
