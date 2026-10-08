package bot

import (
	"strings"
	"testing"

	"github.com/azzimoda/rubix"
)

func TestParseScrambleArgNoArg(t *testing.T) {
	moves, err := parseScrambleArg("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if moves != nil {
		t.Fatalf("expected nil moves for empty arg, got %v", moves)
	}
}

func TestParseScrambleArgMoveString(t *testing.T) {
	moves, err := parseScrambleArg("R U F' L B2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(moves) != 5 {
		t.Fatalf("expected 5 moves, got %d (%v)", len(moves), moves)
	}
	if got := movesString(moves); got != "R U F' L B2" {
		t.Fatalf("unexpected move sequence: %q", got)
	}
}

func TestParseScrambleArgLength(t *testing.T) {
	moves, err := parseScrambleArg("40")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(moves) != 40 {
		t.Fatalf("expected 40 moves, got %d", len(moves))
	}
}

func TestParseScrambleArgInvalidNotation(t *testing.T) {
	for _, arg := range []string{"Q", "R &&&", "FZZZ", "(R"} {
		if _, err := parseScrambleArg(arg); err == nil {
			t.Fatalf("expected error for %q", arg)
		}
	}
}

func TestParseScrambleArgLengthOutOfRange(t *testing.T) {
	for _, arg := range []string{"0", "-5", "1001"} {
		if _, err := parseScrambleArg(arg); err == nil {
			t.Fatalf("expected error for length %q", arg)
		}
	}
}

func TestParseScrambleArgRejectsSolvedResult(t *testing.T) {
	// "R R'" cancels itself out and leaves the cube solved.
	if _, err := parseScrambleArg("R R'"); err == nil {
		t.Fatalf("expected error for a scramble that leaves the cube solved")
	}
}

func TestCommandArgument(t *testing.T) {
	cases := []struct{ text, command, want string }{
		{"/start", "start", ""},
		{"/start   ", "start", ""},
		{"/start R U F'", "start", "R U F'"},
		{"/start  R2 ", "start", "R2"},
		{"/start physics", "start", "physics"},
		{"hello world", "start", ""},
	}
	for _, c := range cases {
		if got := commandArgument(c.text, c.command); got != c.want {
			t.Fatalf("commandArgument(%q, %q) = %q, want %q", c.text, c.command, got, c.want)
		}
	}
}

// movesString joins moves by their notation, mirroring model.MovesString.
func movesString(moves []rubix.Move) string {
	parts := make([]string, len(moves))
	for i, m := range moves {
		parts[i] = m.String()
	}
	return strings.Join(parts, " ")
}
