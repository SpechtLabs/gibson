// Package glitch tears a few rows of the finished frame sideways for a
// moment, every few seconds.
package glitch

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Glitch tears a few rows sideways for a moment, every period seconds from
// some time on.
type Glitch struct {
	from, delay, period float64
}

// Option configures a [Glitch].
type Option func(*Glitch)

// WithFrom sets when the glitch cycle starts counting (default 7s).
func WithFrom(t float64) Option { return func(g *Glitch) { g.from = t } }

// WithDelay skips the first d seconds of the cycle (default 1s).
func WithDelay(d float64) Option { return func(g *Glitch) { g.delay = d } }

// WithPeriod sets the cycle's period (default 4.3s).
func WithPeriod(p float64) Option { return func(g *Glitch) { g.period = p } }

// New returns a glitch.
func New(opts ...Option) *Glitch {
	g := &Glitch{from: 7, delay: 1, period: 4.3}
	for _, o := range opts {
		o(g)
	}
	return g
}

// Apply tears the composed frame f.
func (g *Glitch) Apply(st *stage.Stage, f *r.Frame) {
	t := st.T
	if t < g.from+g.delay {
		return
	}
	cyc := math.Floor((t - g.from) / g.period)
	local := math.Mod(t-g.from, g.period)
	if local > 0.16 {
		return
	}
	k := uint32(cyc)
	step := uint32(local * 30)
	for b := range 3 {
		y0 := int(anim.Hash(k, uint32(b), 21) * float64(f.H))
		bh := 1 + int(anim.Hash(k, uint32(b), 22)*3)
		shift := int((anim.Hash(k, uint32(b), 23+step) - 0.5) * 14)
		for y := y0; y < min(f.H, y0+bh); y++ {
			row := make([]r.Cell, f.W)
			copy(row, f.Cells[y*f.W:(y+1)*f.W])
			for x := range f.W {
				src := row[((x-shift)%f.W+f.W)%f.W]
				src.FG = src.FG.Mul(r.RGB{R: 0.6, G: 1.2, B: 1.3})
				src.BG = src.BG.Add(st.Palette.Grad[2].Scale(0.06))
				f.Cells[y*f.W+x] = src
			}
		}
	}
}
