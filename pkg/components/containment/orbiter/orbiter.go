// Package orbiter is a containment field like an atom model: nodes circle
// the specimen on tilted paths, each trailing a fading comet tail.
package orbiter

import (
	"math"

	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Field is the orbiter field.
type Field struct {
	orbits []orbit
	radius float64
}

type orbit struct{ tiltX, tiltZ, speed, phase float64 }

// Option configures a [Field].
type Option func(*Field)

// WithRadius sets the orbits' radius (default 1.35).
func WithRadius(rad float64) Option { return func(o *Field) { o.radius = rad } }

// New returns an orbiter field.
func New(opts ...Option) *Field {
	o := &Field{
		orbits: []orbit{{1.15, 0.25, 1.6, 0}, {-0.95, 0.9, -1.25, 2.1}, {0.2, -1.2, 1.05, 4.2}},
		radius: 1.35,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Name implements [containment.Field].
func (*Field) Name() string { return "orbiter" }

// Draw implements [containment.Field].
func (o *Field) Draw(st *stage.Stage, in float64) {
	t := st.T
	const tail, steps = 2.2, 36 // tail length (rad), segments
	for _, ob := range o.orbits {
		rad := o.radius * (0.25 + 0.75*in)
		pt := func(a float64) r.Vec3 {
			return r.V(math.Cos(a)*rad, 0, math.Sin(a)*rad).RotX(ob.tiltX).RotZ(ob.tiltZ).Add(r.V(0, -0.02, 0))
		}
		head := ob.phase + ob.speed*t
		dir := 1.0
		if ob.speed < 0 {
			dir = -1
		}
		p := st.Pen(stage.Identity, stage.Solid(st.Palette.Rings), 0)
		for k := range steps {
			f0, f1 := float64(k)/steps, float64(k+1)/steps
			p.I = 1.4 * in * math.Pow(1-f0, 1.1)
			p.Seg(pt(head-dir*tail*f0), pt(head-dir*tail*f1), 1)
		}
		// the node itself: a short bright tick across the path
		h := pt(head)
		p.I = 1.8 * in
		p.Color = stage.Solid(st.Palette.Rings.Lerp(r.RGB{R: 1, G: 1, B: 1}, 0.5))
		p.Seg(h.Add(r.V(0, 0.06, 0)), h.Sub(r.V(0, 0.06, 0)), 1)
		p.Seg(h.Add(r.V(0.06, 0, 0)), h.Sub(r.V(0.06, 0, 0)), 1)
	}
}
