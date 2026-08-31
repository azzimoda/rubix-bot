package bot

import (
	"strings"
	"testing"

	"github.com/azzimoda/rubix-bot/internal/model"
)

func TestBoardCaption(t *testing.T) {
	sess := &model.Session{
		Scramble: "R U F2",
	}
	got := boardCaption(sess)
	for _, want := range []string{"Scramble: R U F2", "Moves (0)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("caption missing %q:\n%s", want, got)
		}
	}

	sess.AppendMove("R")
	sess.AppendMove("U'")
	if got := boardCaption(sess); !strings.Contains(got, "Moves (2): R U'") {
		t.Fatalf("expected updated moves in caption:\n%s", got)
	}

	sess.Solved = true
	if got := boardCaption(sess); !strings.Contains(got, "Solved") {
		t.Fatalf("expected solved text in caption:\n%s", got)
	}
}

func TestBoardKeyboardHasMoves(t *testing.T) {
	kb := boardKeyboard(false)
	count := 0
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, "move:") {
				count++
			}
		}
	}
	if count != 18 {
		t.Fatalf("expected 18 move buttons, got %d", count)
	}
	for _, want := range []string{"move:x", "move:x'", "move:y", "move:y'", "move:z", "move:z'"} {
		found := false
		for _, row := range kb.InlineKeyboard {
			for _, btn := range row {
				if btn.CallbackData == want {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("missing axis button %q", want)
		}
	}
}

func TestBoardKeyboardSolvedOnlyRestart(t *testing.T) {
	kb := boardKeyboard(true)
	moveCount := 0
	restartCount := 0
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, "move:") {
				moveCount++
			}
			if btn.CallbackData == "session:restart" {
				restartCount++
			}
		}
	}
	if moveCount != 0 {
		t.Fatalf("expected 0 move buttons when solved, got %d", moveCount)
	}
	if restartCount != 1 {
		t.Fatalf("expected exactly 1 restart button when solved, got %d", restartCount)
	}
}
