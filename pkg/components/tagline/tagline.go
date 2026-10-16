// Package tagline is the line typed out under the title, with a blinking
// block cursor.
package tagline

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/stage"
)

// Tagline is the tagline.
type Tagline struct {
	text  []rune
	start float64
	cps   float64
}

// Option configures a [Tagline].
type Option func(*Tagline)

// WithStart sets when typing starts (default 6s).
func WithStart(t float64) Option { return func(tl *Tagline) { tl.start = t } }

// New returns a tagline reading text.
func New(text string, opts ...Option) *Tagline {
	tl := &Tagline{text: []rune(text), start: 6, cps: 38}
	for _, o := range opts {
		o(tl)
	}
	return tl
}

// Draw implements [stage.Layer].
func (tl *Tagline) Draw(st *stage.Stage) {
	t := st.T
	if t < tl.start {
		return
	}
	col := st.Palette.Text.Scale(0.85)
	n := min(len(tl.text), int((t-tl.start)*tl.cps))
	ox := (st.Layout.W - len(tl.text)) / 2
	st.Canvas.Text(ox, st.Layout.TagY, string(tl.text[:n]), col, 0.25)
	if math.Mod(t, 0.9) < 0.5 {
		st.Canvas.SetCell(ox+n, st.Layout.TagY, '█', col, 0.3)
	}
}
