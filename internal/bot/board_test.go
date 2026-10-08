package bot

import (
	"strings"
	"testing"

	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/go-telegram/bot/models"
)

// findBlocks returns all blocks of the given type in the rich message.
func findBlocks(rich *models.InputRichMessage, t models.RichBlockType) []models.InputRichBlock {
	var out []models.InputRichBlock
	for _, b := range rich.Blocks {
		if b.Type == t {
			out = append(out, b)
		}
	}
	return out
}

func blockTexts(blocks []models.InputRichBlock) []string {
	var out []string
	for _, b := range blocks {
		switch b.Type {
		case models.RichBlockTypeSectionHeading:
			out = append(out, b.InputRichBlockSectionHeading.Text.PlainText)
		case models.RichBlockTypePreformatted:
			out = append(out, b.InputRichBlockPreformatted.Text.PlainText)
		}
	}
	return out
}

func TestBoardRichMessageActive(t *testing.T) {
	sess := &model.Session{Scramble: "R U F2"}
	sess.AppendMove("R")
	sess.AppendMove("U'")
	sess.State = testSolvedState()
	rich, err := boardRichMessage(sess)
	if err != nil {
		t.Fatalf("boardRichMessage: %v", err)
	}

	// Move log in a details block: summary counts them, content lists them.
	details := findBlocks(rich, models.RichBlockTypeDetails)
	if len(details) != 1 {
		t.Fatalf("expected 1 details block, got %d", len(details))
	}
	if !details[0].InputRichBlockDetails.IsOpen {
		t.Fatalf("active game details should be open")
	}
	if !strings.Contains(details[0].InputRichBlockDetails.Summary.PlainText, "Moves (2)") {
		t.Fatalf("unexpected summary: %q", details[0].InputRichBlockDetails.Summary.PlainText)
	}
	logs := blockTexts(details[0].InputRichBlockDetails.Blocks)
	if len(logs) != 1 || logs[0] != "R U'" {
		t.Fatalf("unexpected move log: %v", logs)
	}

	// Buttons: 18 moves + restart while active.
	buttons := findBlocks(rich, models.RichBlockTypeButtons)
	if len(buttons) != 1 {
		t.Fatalf("expected 1 buttons block, got %d", len(buttons))
	}
	btnList := buttons[0].InputRichBlockButtons.Buttons
	moveCount, restartCount := 0, 0
	for _, btn := range btnList {
		if strings.HasPrefix(btn.CallbackData, "move:") {
			moveCount++
		}
		if btn.CallbackData == "session:restart" {
			restartCount++
		}
	}
	if moveCount != 18 {
		t.Fatalf("expected 18 move buttons, got %d", moveCount)
	}
	if restartCount != 1 {
		t.Fatalf("expected exactly 1 restart button, got %d", restartCount)
	}
	for _, want := range []string{"move:x", "move:x'", "move:y", "move:y'", "move:z", "move:z'"} {
		found := false
		for _, btn := range btnList {
			if btn.CallbackData == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing axis button %q", want)
		}
	}
}

func TestBoardRichMessageSolved(t *testing.T) {
	sess := &model.Session{Scramble: "R U F2", Solved: true, Moves: "R U'"}
	sess.State = testSolvedState()
	rich, err := boardRichMessage(sess)
	if err != nil {
		t.Fatalf("boardRichMessage: %v", err)
	}

	headings := findBlocks(rich, models.RichBlockTypeSectionHeading)
	if len(headings) != 1 || !strings.Contains(headings[0].InputRichBlockSectionHeading.Text.PlainText, "Solved") {
		t.Fatalf("expected solved heading, got %+v", headings)
	}

	// Move log must be collapsed once solved.
	details := findBlocks(rich, models.RichBlockTypeDetails)
	if len(details) != 1 {
		t.Fatalf("expected 1 details block, got %d", len(details))
	}
	if details[0].InputRichBlockDetails.IsOpen {
		t.Fatalf("solved game details should be collapsed")
	}

	// Only restart remains once solved.
	buttons := findBlocks(rich, models.RichBlockTypeButtons)
	if len(buttons) != 1 {
		t.Fatalf("expected 1 buttons block, got %d", len(buttons))
	}
	btnList := buttons[0].InputRichBlockButtons.Buttons
	moveCount, restartCount := 0, 0
	for _, btn := range btnList {
		if strings.HasPrefix(btn.CallbackData, "move:") {
			moveCount++
		}
		if btn.CallbackData == "session:restart" {
			restartCount++
		}
	}
	if moveCount != 0 {
		t.Fatalf("expected 0 move buttons when solved, got %d", moveCount)
	}
	if restartCount != 1 {
		t.Fatalf("expected exactly 1 restart button when solved, got %d", restartCount)
	}
}

func TestValidMoveMatchesRichButtons(t *testing.T) {
	valid := []string{"F", "F'", "B", "B'", "U", "U'", "D", "D'", "L", "L'", "R", "R'", "x", "x'", "y", "y'", "z", "z'"}
	for _, m := range valid {
		if !validMove(m) {
			t.Fatalf("expected %q to be a valid move", m)
		}
	}
	invalid := []string{"", "Q", "X", "M", "F2", "R2", "u", "l", "b", "d", "f", "r"}
	for _, m := range invalid {
		if validMove(m) {
			t.Fatalf("expected %q to be an invalid move", m)
		}
	}
}

// testSolvedState returns the serialized state of a freshly solved 3x3 cube.
func testSolvedState() string {
	return "000 000 000\n111 111 111\n222 222 222\n333 333 333\n444 444 444\n555 555 555"
}
