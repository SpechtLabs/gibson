// Package projector is the hologram projector under the specimen: an
// emitter pad on the vessel's floor, rings of light on the floor around it,
// a beam, and the scan that materializes a specimen from the bottom up.
//
// The projector draws any [specimen.Specimen]; it owns how a specimen is
// shown (its sway, bob and scan), the specimen only what it looks like.
package projector

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/components/specimen"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Projector is the projector.
type Projector struct {
	floor float64
	rings []floorRing
}

type floorRing struct{ y, rad, speed, span, ph float64 }

// Option configures a [Projector].
type Option func(*Projector)

// WithFloor sets how far below the center the projector stands (default
// 1.05): the vessel's [vessel.Vessel.Floor].
func WithFloor(f float64) Option { return func(p *Projector) { p.floor = f } }

// New returns a projector.
func New(opts ...Option) *Projector {
	p := &Projector{
		floor: 1.05,
		rings: []floorRing{
			{-1.12, 1.5, 0.35, 1.7, 0}, {-1.12, 1.8, -0.22, 1.25, 1.3}, {-1.14, 2.15, 0.15, 1.55, 2.2},
		},
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// place is how the projector shows a specimen: scaled, swaying and bobbing.
func place(st *stage.Stage, spec specimen.Specimen) func(r.Vec3) r.Vec3 {
	t := st.T
	sway := 0.45*math.Sin(t*0.45) + 0.1*math.Sin(t*1.3)
	bob := 0.05 * math.Sin(t*1.7)
	scale := spec.Scale()
	return func(v r.Vec3) r.Vec3 {
		return v.Scale(scale).RotY(sway + math.Pi*0.08).Add(r.V(0, -0.02+bob, 0))
	}
}

// scanLevel is the model-space height the scan has reached.
func scanLevel(reveal float64) float64 { return -1.0 + reveal*2.05 }

// Materialize draws spec up to the scan height reveal (0 to 1), noisy near
// the scan front. It records the specimen's depth in the canvas, so what's
// drawn after it can be tucked behind it.
func (*Projector) Materialize(st *stage.Stage, spec specimen.Specimen, reveal float64) {
	t := st.T
	level := scanLevel(reveal)
	xf := place(st, spec)
	pulse := 1.1 + 0.12*math.Sin(t*3.1)
	pal := st.Palette
	p := st.Pen(xf, nil, pulse)
	specCol := stage.Solid(pal.Specimen)
	q := uint32(t * 20) // jitter changes 20×/s
	st.Canvas.MarkMask = true
	defer func() { st.Canvas.MarkMask = false }()
	for li, sd := range spec.Strands(t) {
		pl := sd.Pts
		p.I = pulse * sd.I
		p.Color = specCol
		if sd.Accent {
			p.Color = stage.Solid(pal.Eye)
		}
		for i := 0; i+1 < len(pl); i++ {
			a, b := pl[i], pl[i+1]
			if a.Y > level && b.Y > level {
				continue
			}
			// clip at the scan level
			if b.Y > level {
				b = a.Lerp(b, (level-a.Y)/(b.Y-a.Y))
			} else if a.Y > level {
				a = b.Lerp(a, (level-b.Y)/(a.Y-b.Y))
			}
			// noisy near the scan front
			if reveal < 1 && level-a.Y < 0.15 {
				j := (anim.Hash(uint32(li), uint32(i), q) - 0.5) * 0.12
				a.X += j
				b.X += j
			}
			p.Seg(a, b, 1)
		}
	}
}

// ScanRing draws the ring of light at the scan front, until the scan is
// done.
func (*Projector) ScanRing(st *stage.Stage, spec specimen.Specimen, reveal float64) {
	if reveal >= 1 {
		return
	}
	rp := st.Pen(st.Turntable, stage.Solid(st.Palette.Beam), 1.3)
	y := place(st, spec)(r.V(0, scanLevel(reveal), 0)).Y
	rp.Circle(y, 1.0, 48)
}

// Beam draws the projector's flickering beam at power (0 to 1), dimmed
// while the sequence is idle.
func (p *Projector) Beam(st *stage.Stage, power float64, idle bool) {
	q := uint32(st.T * 20)
	bi := power * (0.55 + 0.25*anim.Hash(q, 77, 1))
	if idle {
		bi *= 0.6
	}
	if bi <= 0 {
		return
	}
	bp := st.Pen(func(v r.Vec3) r.Vec3 { return v.RotY(-st.Yaw * 0.5) }, stage.Solid(st.Palette.Beam), bi)
	src := r.V(0, -p.floor+0.02, 0)
	for k := range 6 {
		a := float64(k) / 6 * 2 * math.Pi
		bp.Seg(src, r.V(math.Cos(a)*0.3, -0.3, math.Sin(a)*0.3), 1)
	}
}

// FloorRings draws the arcs of light circling on the floor, in from 0 to 1.
func (p *Projector) FloorRings(st *stage.Stage, in float64) {
	if in <= 0 {
		return
	}
	t := st.T
	pen := st.Pen(stage.Identity, stage.Solid(st.Palette.Orbit), 0.6)
	for _, rg := range p.rings {
		const seg = 64
		start := rg.ph + rg.speed*t
		span := rg.span * math.Pi * in
		rad := rg.rad * (0.6 + 0.4*in)
		pts := make([]r.Vec3, 0, seg+1)
		for k := 0; k <= seg; k++ {
			a := start + span*float64(k)/seg
			pts = append(pts, r.V(math.Cos(a)*rad, rg.y, math.Sin(a)*rad))
		}
		pen.Poly(pts, 1)
	}
}

// Pad draws the hatched emitter plate the specimen stands on, the first
// build of it drawn.
func (p *Projector) Pad(st *stage.Stage, build float64) {
	vp := st.Pen(st.Turntable, st.HeightColor, 0.45)
	for k := range 6 {
		z := -0.32 + float64(k)*0.128
		vp.Seg(r.V(-0.45, -p.floor+0.02, z), r.V(0.45, -p.floor+0.02, z), build)
	}
}
