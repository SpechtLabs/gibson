// Package banner is a big red alert panel stamped over a finished frame: a
// heading in block letters (or spaced capitals when the terminal is too
// narrow for them), a message, and an attribution, with the frame behind it
// dimmed so the panel pops.
package banner

import (
	"strings"

	"github.com/spechtlabs/gibson/pkg/font"
	r "github.com/spechtlabs/gibson/pkg/render"
)

// Banner is the panel.
type Banner struct {
	heading, message, attribution string
	panel, border, text, fade     r.RGB
	dim                           float64
}

// Option configures a [Banner].
type Option func(*Banner)

// WithAttribution sets a dim line under the message.
func WithAttribution(s string) Option { return func(b *Banner) { b.attribution = s } }

// WithColors sets the panel's background, border and text colors (default
// alarm red with white text).
func WithColors(panel, border, text r.RGB) Option {
	return func(b *Banner) { b.panel, b.border, b.text = panel, border, text }
}

// WithDim sets how bright the frame behind the panel stays (default 0.4).
func WithDim(f float64) Option { return func(b *Banner) { b.dim = f } }

// New returns a banner with a heading and a message.
func New(heading, message string, opts ...Option) *Banner {
	b := &Banner{
		heading: heading,
		message: message,
		panel:   r.Hex(0x9c0016),
		border:  r.Hex(0xff2a36),
		text:    r.Hex(0xffffff),
		fade:    r.Hex(0xffb3b8),
		dim:     0.4,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Apply dims f and stamps the panel over its middle.
func (b *Banner) Apply(f *r.Frame) {
	for i := range f.Cells {
		f.Cells[i].FG = f.Cells[i].FG.Scale(b.dim)
		f.Cells[i].BG = f.Cells[i].BG.Scale(b.dim)
	}

	const gap, word = 2, 3 // pixels between letters, and for a space
	blockW := 0
	for _, ch := range b.heading {
		if ch == ' ' {
			blockW += word
		} else {
			blockW += font.Width + gap
		}
	}
	blockW -= gap
	inner := f.W - 8 // panel border and padding, and a margin around it
	blocks := blockW <= inner && f.H >= 16

	headRows := 1
	if blocks {
		headRows = (font.Height + 1) / 2
	}
	lines := wrap(b.message, inner)
	rows := 1 + headRows + 1 + len(lines) + 1 // padding, heading, gap, message, padding
	if b.attribution != "" {
		rows++
	}
	width := min(inner, max(blockW, len([]rune(b.heading))*2, longest(lines), len([]rune(b.attribution))))
	x0 := (f.W-width)/2 - 2
	y0 := max(0, (f.H-rows-2)/2)
	b.box(f, x0, y0, width+4, rows+2)

	y := y0 + 2
	if blocks {
		b.blockText(f, (f.W-blockW)/2, y)
	} else {
		spaced := strings.Join(strings.Split(b.heading, ""), " ")
		b.text1(f, (f.W-len([]rune(spaced)))/2, y, spaced, b.text)
	}
	y += headRows + 1
	for _, l := range lines {
		b.text1(f, (f.W-len([]rune(l)))/2, y, l, b.text)
		y++
	}
	if b.attribution != "" {
		b.text1(f, (f.W-len([]rune(b.attribution)))/2, y, b.attribution, b.fade)
	}
}

// box fills the panel and draws its border.
func (b *Banner) box(f *r.Frame, x0, y0, w, h int) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			ch := ' '
			switch {
			case y == y0 && x == x0:
				ch = '┏'
			case y == y0 && x == x0+w-1:
				ch = '┓'
			case y == y0+h-1 && x == x0:
				ch = '┗'
			case y == y0+h-1 && x == x0+w-1:
				ch = '┛'
			case y == y0 || y == y0+h-1:
				ch = '━'
			case x == x0 || x == x0+w-1:
				ch = '┃'
			}
			b.set(f, x, y, ch, b.border, b.panel)
		}
	}
}

// blockText draws the heading in block letters, two pixel rows per cell.
func (b *Banner) blockText(f *r.Frame, x0, y0 int) {
	const gap, word = 2, 3
	x := x0
	for _, ch := range b.heading {
		if ch == ' ' {
			x += word
			continue
		}
		g := font.Get(ch)
		for cy := 0; cy*2 < font.Height; cy++ {
			for gx := range font.Width {
				top := g.Lit(gx, cy*2)
				bot := cy*2+1 < font.Height && g.Lit(gx, cy*2+1)
				switch {
				case top && bot:
					b.set(f, x+gx, y0+cy, '█', b.text, b.panel)
				case top:
					b.set(f, x+gx, y0+cy, '▀', b.text, b.panel)
				case bot:
					b.set(f, x+gx, y0+cy, '▄', b.text, b.panel)
				}
			}
		}
		x += font.Width + gap
	}
}

func (b *Banner) text1(f *r.Frame, x, y int, s string, col r.RGB) {
	for i, ch := range []rune(s) {
		b.set(f, x+i, y, ch, col, b.panel)
	}
}

func (b *Banner) set(f *r.Frame, x, y int, ch rune, fg, bg r.RGB) {
	if x < 0 || y < 0 || x >= f.W || y >= f.H {
		return
	}
	f.Cells[y*f.W+x] = r.Cell{R: ch, FG: fg, BG: bg}
}

// wrap breaks s into lines at most w wide, at spaces.
func wrap(s string, w int) []string {
	var lines []string
	line := ""
	for word := range strings.FieldsSeq(s) {
		switch {
		case line == "":
			line = word
		case len([]rune(line))+1+len([]rune(word)) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	return append(lines, line)
}

func longest(lines []string) int {
	n := 0
	for _, l := range lines {
		n = max(n, len([]rune(l)))
	}
	return n
}
