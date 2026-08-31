// Package render turns a rubix.Cube into a single composite PNG suitable for
// sending through Telegram. The board shows the two orthographic perspective
// views (White-Green-Red and Yellow-Orange-Blue) side by side, with the
// unfolded net overlaid in the bottom-right corner.
package render

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"strings"

	"github.com/azzimoda/rubix"
	"github.com/azzimoda/rubix/render"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// const-ish layout parameters.
var (
	// perspectiveSize roughly corresponds to the printed sticker size in pixels,
	// which sets how large the two perspective views are rendered.
	perspectiveSize = 90.0
	netScale        = 0.42
)

// Board renders the composite board image for the cube and returns its PNG
// bytes.
func Board(c *rubix.Cube) ([]byte, error) {
	wgr := render.PerspectiveSVG(c, "wgr", perspectiveSize)
	yob := render.PerspectiveSVG(c, "yob", perspectiveSize)
	net := render.NetSVG(c, 24)

	wgrImg, err := svgToImage(wgr)
	if err != nil {
		return nil, fmt.Errorf("render wgr perspective: %w", err)
	}
	yobImg, err := svgToImage(yob)
	if err != nil {
		return nil, fmt.Errorf("render yob perspective: %w", err)
	}
	netImg, err := svgToImage(net)
	if err != nil {
		return nil, fmt.Errorf("render net: %w", err)
	}

	return composite(wgrImg, yobImg, netImg)
}

// svgToImage parses an SVG string and renders it to an RGBA image of exactly
// its viewBox dimensions.
func svgToImage(svg string) (*image.RGBA, error) {
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg))
	if err != nil {
		return nil, err
	}

	w := int(icon.ViewBox.W)
	h := int(icon.ViewBox.H)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid viewBox %dx%d", w, h)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.Draw(dasher, 1.0)
	return img, nil
}

// composite lays the two perspective images side by side and stamps the net in
// the bottom-right corner.
func composite(wgr, yob, net *image.RGBA) ([]byte, error) {
	perspW := wgr.Bounds().Dx()
	perspH := wgr.Bounds().Dy()

	gap := 8
	canvasW := perspW*2 + gap
	canvasH := perspH

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	// Perspectives side by side.
	draw.Draw(canvas, image.Rect(0, 0, perspW, perspH), wgr, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(perspW+gap, 0, perspW+gap+perspW, perspH), yob, image.Point{}, draw.Src)

	// Net overlaid in the bottom-right corner.
	netW := int(float64(net.Bounds().Dx()) * netScale)
	netH := int(float64(net.Bounds().Dy()) * netScale)
	netImg := scaleTo(net, netW, netH)
	x := canvasW - netW - gap
	y := canvasH - netH - gap
	draw.Draw(canvas, image.Rect(x, y, x+netW, y+netH), netImg, image.Point{}, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// scaleTo returns a nearest-neighbour scaled copy of src to the given size.
func scaleTo(src *image.RGBA, dstW, dstH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			sx := x * srcW / dstW
			sy := y * srcH / dstH
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
