// Package poweron is an old tube switching on: a horizontal line snapping
// open, then the picture blooming out from it vertically.
package poweron

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// PowerOn is a horizontal line snapping open, then the picture blooming
// vertically.
type PowerOn struct {
	snap, bloom float64
}

// Option configures a [PowerOn].
type Option func(*PowerOn)

// WithSnap sets how long the line takes to snap open (default 0.22s).
func WithSnap(d float64) Option { return func(p *PowerOn) { p.snap = d } }

// WithBloom sets how long the picture then takes to bloom (default
// 0.33s).
func WithBloom(d float64) Option { return func(p *PowerOn) { p.bloom = d } }

// New returns a power-on.
func New(opts ...Option) *PowerOn {
	p := &PowerOn{snap: 0.22, bloom: 0.33}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Draw implements [stage.Layer].
func (p *PowerOn) Draw(st *stage.Stage) {
	t := st.T
	if t >= p.snap+p.bloom {
		return
	}
	w, h := st.Layout.W, st.Layout.H
	glow := st.Palette.Grad[1]
	white := r.RGB{R: 1, G: 1, B: 1}
	mid := h / 2
	if t < p.snap {
		half := int(anim.Ease(t/p.snap) * float64(w) / 2)
		for x := w/2 - half; x < w/2+half; x++ {
			st.Canvas.SetCell(x, mid, '━', white, 1.5)
			st.Canvas.AddBG(x, mid, glow.Scale(0.35))
		}
		return
	}
	ph := (t - p.snap) / p.bloom
	half := anim.Ease(ph) * float64(h) / 2
	for y := range h {
		d := math.Abs(float64(y - mid))
		if d > half {
			continue
		}
		v := (1 - ph) * (1 - d/(half+1)) * 0.7
		for x := range w {
			st.Canvas.AddBG(x, y, white.Lerp(glow, 0.5).Scale(v))
		}
	}
}
