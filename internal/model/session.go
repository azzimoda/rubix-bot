package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/azzimoda/rubix"
	"gorm.io/gorm"
)

// ScrambleLength is the number of random moves used when starting a session.
const ScrambleLength = 20

// NewSession creates a new session with a freshly scrambled cube for the given
// chat. The scramble uses a time-based seed so consecutive sessions differ.
func NewSession(chatID uint) *Session {
	return NewSessionFor(chatID, rubix.Scramble(ScrambleLength, rubix.NewRandom(time.Now().UnixNano())))
}

// NewSessionFor creates a new session whose cube is scrambled by the given
// moves. The scramble is recorded verbatim so the player can review it.
func NewSessionFor(chatID uint, scramble []rubix.Move) *Session {
	c := rubix.New(3)
	c.ApplyMoves(scramble)

	return &Session{
		ChatID:   chatID,
		State:    marshalCube(c),
		Scramble: MovesString(scramble),
		Moves:    "",
		Solved:   false,
		Ended:    false,
	}
}

type Session struct {
	gorm.Model

	// ChatID references the owning chat.
	ChatID uint `gorm:"column:chat_id;index:idx_sessions_chat_ended,priority:1"`
	// State is the serialized cube.
	State string `gorm:"column:state;not null"`
	// Scramble is the initial scramble applied to the cube.
	Scramble string `gorm:"column:scramble;not null"`
	// Moves is the space-separated log of moves made by the user so far.
	Moves string `gorm:"column:moves;not null;default:''"`
	// Solved is true once the cube has been solved.
	Solved bool `gorm:"column:solved;not null;default:false"`
	// Ended is true once the session is finished (solved or abandoned).
	Ended bool `gorm:"column:ended;not null;default:false;index:idx_sessions_chat_ended,priority:2"`
}

func (s *Session) SetCube(c *rubix.Cube)      { s.State = marshalCube(c) }
func (s *Session) Cube() (*rubix.Cube, error) { return unmarshalCube(s.State) }

// AppendMove appends a move name to the session's move log.
func (s *Session) AppendMove(name string) {
	if s.Moves == "" {
		s.Moves = name
		return
	}
	s.Moves = s.Moves + " " + name
}

// MovesList returns the logged moves as a slice of move names.
func (s *Session) MovesList() []string {
	if strings.TrimSpace(s.Moves) == "" {
		return nil
	}
	return strings.Fields(s.Moves)
}

// MovesString joins the given moves into a space-separated string.
func MovesString(moves []rubix.Move) string {
	parts := make([]string, len(moves))
	for i, m := range moves {
		parts[i] = m.String()
	}
	return strings.Join(parts, " ")
}

func marshalCube(cube *rubix.Cube) string {
	var b strings.Builder
	for f := 0; f < 6; f++ {
		if f != 0 {
			b.WriteString("\n")
		}
		face := cube.Face(f)
		for r, row := range face {
			if r != 0 {
				b.WriteString(" ")
			}
			for _, c := range row {
				fmt.Fprintf(&b, "%d", c)
			}
		}
	}
	return strings.TrimSpace(b.String())
}
func unmarshalCube(s string) (*rubix.Cube, error) {
	facesStr := strings.Split(s, "\n")
	if len(facesStr) != 6 {
		return nil, fmt.Errorf("wrong number of faces")
	}

	n := 0
	// faces := [6][][]uint8{}
	var cube *rubix.Cube
	for f, faceStr := range facesStr {
		rowsStr := strings.Split(faceStr, " ")
		nRows := len(rowsStr)
		if nRows < 2 {
			return nil, fmt.Errorf("wrong number of rows: less than 2")
		}
		if n == 0 {
			n = nRows
			cube = rubix.New(n)
		} else if nRows != n {
			return nil, fmt.Errorf("different number of rows")
		}

		columns := len(rowsStr[0])
		if columns != n {
			return nil, fmt.Errorf("face is not a square: different number of rows and columns")
		}

		for r, row := range rowsStr {
			if len(row) != n {
				return nil, fmt.Errorf("face is not a square: row %d has %d columns", r, len(row))
			}
			for c, color := range row {
				colorCode := uint8(color - '0')
				if colorCode > 5 {
					return nil, fmt.Errorf("invalid color code: %d", colorCode)
				}

				cube.Set(f, r, c, colorCode)
			}
		}
	}

	return cube, nil
}
