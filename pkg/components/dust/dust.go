// Package dust is the ambient drift around the scene: motes and short
// streaks rising slowly and twinkling.
package dust

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Dust is a field of particles, fixed when it's made.
type Dust struct {
	parts []particle
}

type particle struct {
	pos   r.Vec3
	dir   r.Vec3
	line  bool
	speed float64
	phase float64
}

// Option configures [Dust].
type Option func(*options)

type options struct{ count int }

// WithCount sets how many particles there are (default 70).
func WithCount(n int) Option { return func(o *options) { o.count = n } }

// New returns a dust field. Particles are placed by hash, so every field of
// the same size is the same.
func New(opts ...Option) *Dust {
	o := options{count: 70}
	for _, opt := range opts {
		opt(&o)
	}
	d := &Dust{parts: make([]particle, 0, o.count)}
	for i := range o.count {
		a := anim.Hash(uint32(i), 1, 0) * 2 * math.Pi
		rad := 1.5 + anim.Hash(uint32(i), 2, 0)*2.4
		d.parts = append(d.parts, particle{
			pos:   r.V(math.Cos(a)*rad, -1.3+anim.Hash(uint32(i), 3, 0)*3.4, math.Sin(a)*rad),
			dir:   r.V(anim.Hash(uint32(i), 4, 0)-0.5, anim.Hash(uint32(i), 5, 0)-0.5, anim.Hash(uint32(i), 6, 0)-0.5).Norm(),
			line:  anim.Hash(uint32(i), 7, 0) < 0.55,
			speed: 0.05 + anim.Hash(uint32(i), 8, 0)*0.12,
			phase: anim.Hash(uint32(i), 9, 0) * 10,
		})
	}
	return d
}

// Draw draws the field, faded in from 0 to 1.
func (d *Dust) Draw(st *stage.Stage, in float64) {
	if in <= 0 {
		return
	}
	t := st.T
	pal := st.Palette
	yaw := st.Yaw * 0.35
	for _, pt := range d.parts {
		pos := pt.pos
		pos.Y = math.Mod(pos.Y+1.3+t*pt.speed, 3.4) - 1.3
		pos = pos.RotY(yaw)
		tw := 0.5 + 0.5*math.Sin(t*2.3+pt.phase)
		if pt.line {
			dir := pt.dir.RotY(t * 0.4 * pt.speed * 8)
			p := st.Pen(stage.Identity, stage.Solid(pal.Beam.Lerp(pal.Orbit, 0.35)), 0.4*in*(0.5+tw))
			p.Seg(pos.Sub(dir.Scale(0.12)), pos.Add(dir.Scale(0.12)), 1)
			continue
		}
		w := st.View(pos)
		if x, y, ok := st.Camera.Project(w); ok {
			st.Canvas.Plot(int(x), int(y), w.Z, pal.Rings, 1.2*in*tw)
			st.Canvas.Plot(int(x)+1, int(y), w.Z, pal.Rings, 0.6*in*tw)
		}
	}
}
