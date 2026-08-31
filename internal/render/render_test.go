package render

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/azzimoda/rubix"
)

func TestBoardRendersPNG(t *testing.T) {
	cube := rubix.New(3)
	cube.ApplyMoves([]rubix.Move{rubix.MoveR, rubix.MoveU, rubix.MoveRp, rubix.MoveUp})

	got, err := Board(cube)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		t.Fatalf("bad dimensions %dx%d", cfg.Width, cfg.Height)
	}

	img, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b := img.Bounds(); b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("empty image")
	}
	// Ensure the image is not entirely blank.
	if allOneColor(img) {
		t.Fatalf("image is blank")
	}
	t.Logf("rendered %dx%d board", cfg.Width, cfg.Height)
}

func allOneColor(img image.Image) bool {
	first := img.At(0, 0)
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.At(x, y) != first {
				return false
			}
		}
	}
	return true
}
