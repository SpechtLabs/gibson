// Package preview rasterises frames into images so the demo can be inspected
// (and shared) without a terminal.
package preview

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	r "github.com/spechtlabs/gibson/internal/render"
)

// CW and CH are the size in pixels of one terminal cell in a preview.
const CW, CH = 8, 16

func rgba(c r.RGB) color.RGBA {
	R, G, B := c.Bytes()
	return color.RGBA{R, G, B, 255}
}

// Image renders a frame with an 8x16 cell size.
func Image(f *r.Frame) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, f.W*CW, f.H*CH))
	d := &font.Drawer{Dst: img, Face: basicfont.Face7x13}
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			c := f.Cells[y*f.W+x]
			x0, y0 := x*CW, y*CH
			fill(img, x0, y0, CW, CH, rgba(c.BG))
			fg := rgba(c.FG)
			switch ch := c.R; {
			case ch == ' ':
			case ch >= 0x2800 && ch <= 0x28ff:
				bits := ch - 0x2800
				dots := [8][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {0, 3}, {1, 3}}
				for b, p := range dots {
					if bits&(1<<b) != 0 {
						fill(img, x0+1+p[0]*4, y0+1+p[1]*4, 3, 3, fg)
					}
				}
			case ch == '█':
				fill(img, x0, y0, CW, CH, fg)
			case ch == '▀':
				fill(img, x0, y0, CW, CH/2, fg)
			case ch == '▄':
				fill(img, x0, y0+CH/2, CW, CH/2, fg)
			case ch == '─' || ch == '━':
				fill(img, x0, y0+CH/2-1, CW, 1+btoi(ch == '━'), fg)
			case ch == '│':
				fill(img, x0+CW/2-1, y0, 1, CH, fg)
			case ch == '┌':
				fill(img, x0+CW/2-1, y0+CH/2-1, CW/2+1, 1, fg)
				fill(img, x0+CW/2-1, y0+CH/2-1, 1, CH/2+1, fg)
			case ch == '┐':
				fill(img, x0, y0+CH/2-1, CW/2, 1, fg)
				fill(img, x0+CW/2-1, y0+CH/2-1, 1, CH/2+1, fg)
			case ch == '└':
				fill(img, x0+CW/2-1, y0, 1, CH/2, fg)
				fill(img, x0+CW/2-1, y0+CH/2-1, CW/2+1, 1, fg)
			case ch == '┘':
				fill(img, x0+CW/2-1, y0, 1, CH/2, fg)
				fill(img, x0, y0+CH/2-1, CW/2, 1, fg)
			case ch == '●':
				fill(img, x0+2, y0+5, 4, 6, fg)
			case ch == '▮':
				fill(img, x0+2, y0+3, 4, 10, fg)
			case ch == '▯':
				fill(img, x0+2, y0+3, 4, 1, fg)
				fill(img, x0+2, y0+12, 4, 1, fg)
				fill(img, x0+2, y0+3, 1, 10, fg)
				fill(img, x0+5, y0+3, 1, 10, fg)
			case ch == '▸':
				fill(img, x0+2, y0+6, 4, 4, fg)
			default:
				d.Src = image.NewUniform(fg)
				d.Dot = fixed.P(x0, y0+12)
				d.DrawString(string(ch))
			}
		}
	}
	return img
}

func fill(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), image.NewUniform(c), image.Point{}, draw.Src)
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Palette builds a GIF palette from a set of key colors at many brightness
// levels, plus greys.
func Palette(keys []r.RGB) color.Palette {
	var p color.Palette
	p = append(p, color.RGBA{0, 0, 0, 255})
	for i := 1; i <= 15; i++ {
		v := uint8(i * 17)
		p = append(p, color.RGBA{v, v, v, 255})
	}
	levels := (256 - len(p)) / len(keys)
	for _, k := range keys {
		for l := 1; l <= levels; l++ {
			f := float64(l) / float64(levels)
			p = append(p, rgba(k.Scale(f*f)))
		}
	}
	return p
}

// Quantizer maps preview images onto a fixed palette, caching lookups.
type Quantizer struct {
	pal   color.Palette
	cache map[color.RGBA]uint8
}

// NewQuantizer returns a Quantizer for p.
func NewQuantizer(p color.Palette) *Quantizer {
	return &Quantizer{pal: p, cache: map[color.RGBA]uint8{}}
}

func (q *Quantizer) Paletted(src *image.RGBA) *image.Paletted {
	dst := image.NewPaletted(src.Bounds(), q.pal)
	for i := 0; i < len(src.Pix); i += 4 {
		c := color.RGBA{src.Pix[i], src.Pix[i+1], src.Pix[i+2], 255}
		idx, ok := q.cache[c]
		if !ok {
			idx = uint8(q.pal.Index(c))
			q.cache[c] = idx
		}
		dst.Pix[i/4] = idx
	}
	return dst
}
