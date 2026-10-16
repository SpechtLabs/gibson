// Package ring is a containment field of horizontal rings that drop in
// from above and bob around the specimen.
package ring

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/stage"
)

// Field is the ring field.
type Field struct {
	count  int
	radius float64
}

// Option configures a [Field].
type Option func(*Field)

// WithCount sets how many rings there are (default 3).
func WithCount(n int) Option { return func(rg *Field) { rg.count = n } }

// WithRadius sets the rings' radius (default 0.98).
func WithRadius(rad float64) Option { return func(rg *Field) { rg.radius = rad } }

// New returns a ring field.
func New(opts ...Option) *Field {
	rg := &Field{count: 3, radius: 0.98}
	for _, o := range opts {
		o(rg)
	}
	return rg
}

// Name implements [containment.Field].
func (*Field) Name() string { return "ring" }

// Draw implements [containment.Field].
func (rg *Field) Draw(st *stage.Stage, in float64) {
	t := st.T
	// the rings fade over the specimen in front of it as well as behind
	defer st.Canvas.Occluding(0.18, 0.35)()
	p := st.Pen(stage.Identity, stage.Solid(st.Palette.Rings), 0.9*in)
	for i := range rg.count {
		fi := float64(i)
		base := -0.6 + fi*0.55
		y := base + 0.07*math.Sin(t*1.2+fi*1.7)
		// rings drop in from above
		y += (1 - in) * (1.2 + fi*0.3)
		p.Circle(y, rg.radius+0.03*math.Sin(t*2+fi), 56)
	}
}
