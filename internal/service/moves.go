package service

import "github.com/azzimoda/rubix"

// parseMove parses a single move name (e.g. "R", "R'", "F2") into a rubix.Move.
func parseMove(name string) rubix.Move {
	moves, err := rubix.Parse(name)
	if err != nil || len(moves) == 0 {
		return rubix.MoveNone
	}
	return moves[0]
}
