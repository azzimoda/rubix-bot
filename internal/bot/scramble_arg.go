package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/azzimoda/rubix"
)

const (
	minScrambleLength = 1
	maxScrambleLength = 1000
)

// parseScrambleArg interprets the argument passed to /start. An empty argument
// means "use the default random scramble" and returns nil. A positive integer
// is treated as a scramble length, anything else as a move string.
func parseScrambleArg(arg string) ([]rubix.Move, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return nil, nil
	}

	if n, err := strconv.Atoi(arg); err == nil {
		if n < minScrambleLength || n > maxScrambleLength {
			return nil, fmt.Errorf("scramble length must be between %d and %d, got %d", minScrambleLength, maxScrambleLength, n)
		}
		return rubix.Scramble(n, rubix.NewRandom(time.Now().UnixNano())), nil
	}

	moves, err := rubix.Parse(arg)
	if err != nil {
		return nil, fmt.Errorf("unrecognized scramble %q: %w", arg, err)
	}
	if len(moves) == 0 {
		return nil, fmt.Errorf("empty scramble")
	}
	if isSolvedScramble(moves) {
		return nil, fmt.Errorf("that scramble leaves the cube solved")
	}
	return moves, nil
}

// isSolvedScramble reports whether applying the moves to a fresh cube leaves it
// in the solved state, which would make for a pointless start.
func isSolvedScramble(moves []rubix.Move) bool {
	c := rubix.New(3)
	c.ApplyMoves(moves)
	return c.IsSolved()
}
