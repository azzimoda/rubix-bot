package service

import (
	"context"
	"math"
)

// ChatStats summarizes a chat's ended (finished) sessions for display.
type ChatStats struct {
	// Total is the number of ended sessions.
	Total int
	// Solved is the number of ended sessions that were solved.
	Solved int
	// Abandoned is the number of ended sessions that were left unsolved.
	Abandoned int
	// AvgMovesSolved is the average number of moves across solved sessions.
	// It is 0 when there are no solved sessions.
	AvgMovesSolved int
	// BestMovesSolved is the fewest moves across any solved session. It is 0
	// when there are no solved sessions.
	BestMovesSolved int
}

// SolveRate returns the fraction of ended sessions that were solved, in [0,1].
// It is 0 when there are no ended sessions.
func (s ChatStats) SolveRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Solved) / float64(s.Total)
}

// Stats computes summary statistics over a chat's ended sessions.
func (s *SessionService) Stats(ctx context.Context, chatID uint) (ChatStats, error) {
	ended, err := s.repo.EndedByChat(ctx, chatID)
	if err != nil {
		return ChatStats{}, err
	}

	stats := ChatStats{Total: len(ended)}
	var movesSum, movesMin int
	for i := range ended {
		moves := len(ended[i].MovesList())
		if !ended[i].Solved {
			stats.Abandoned++
			continue
		}
		stats.Solved++
		movesSum += moves
		if movesMin == 0 || moves < movesMin {
			movesMin = moves
		}
	}

	if stats.Solved > 0 {
		stats.AvgMovesSolved = int(math.Round(float64(movesSum) / float64(stats.Solved)))
		stats.BestMovesSolved = movesMin
	}
	return stats, nil
}
