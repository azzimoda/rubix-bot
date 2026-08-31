package model

import (
	"strings"
	"testing"

	"github.com/azzimoda/rubix"
)

func TestCubeMarshalRoundTrip(t *testing.T) {
	original := rubix.New(3)
	original.ApplyMoves([]rubix.Move{rubix.MoveR, rubix.MoveU, rubix.MoveRp, rubix.MoveUp, rubix.MoveF2})

	s := NewSession(1)
	s.SetCube(original)

	recovered, err := s.Cube()
	if err != nil {
		t.Fatalf("Cube: %v", err)
	}
	if recovered.Size() != original.Size() {
		t.Fatalf("size mismatch: %d != %d", recovered.Size(), original.Size())
	}
	for f := 0; f < 6; f++ {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if recovered.At(f, r, c) != original.At(f, r, c) {
					t.Fatalf("sticker mismatch at face=%d row=%d col=%d: %d != %d",
						f, r, c, recovered.At(f, r, c), original.At(f, r, c))
				}
			}
		}
	}
}

func TestNewSessionScramblesCube(t *testing.T) {
	s := NewSession(7)
	c, err := s.Cube()
	if err != nil {
		t.Fatalf("Cube: %v", err)
	}
	if c.IsSolved() {
		t.Fatalf("new session cube should be scrambled")
	}
	if s.Scramble == "" {
		t.Fatalf("expected a scramble string")
	}
	if s.ChatID != 7 {
		t.Fatalf("chat id mismatch: %d", s.ChatID)
	}
}

func TestAppendMove(t *testing.T) {
	s := NewSession(1)
	if got := s.MovesList(); got != nil {
		t.Fatalf("expected nil moves, got %v", got)
	}
	s.AppendMove("R")
	s.AppendMove("U'")
	if got := s.MovesList(); len(got) != 2 || got[0] != "R" || got[1] != "U'" {
		t.Fatalf("unexpected moves: %v", got)
	}
	if s.Moves != "R U'" {
		t.Fatalf("unexpected moves string: %q", s.Moves)
	}
}

func TestUnmarshalCubeRejectsInvalidState(t *testing.T) {
	// Valid state for 6 faces of a solved 3x3 cube.
	face := "000 000 000"
	valid := strings.Repeat(face+"\n", 5) + face
	if _, err := unmarshalCube(valid); err != nil {
		t.Fatalf("valid state rejected: %v", err)
	}

	// A color code of 6 is out of range (faces are 0..5).
	badColor := strings.Replace(valid, "0", "6", 1)
	if _, err := unmarshalCube(badColor); err == nil {
		t.Fatalf("expected error for color code 6")
	}

	// A row that is too long should be rejected.
	long := strings.Replace(valid, "000 000 000", "000 000 0000", 1)
	if _, err := unmarshalCube(long); err == nil {
		t.Fatalf("expected error for non-square face")
	}
}
