// Package hex is a hexagonal prism vessel, a nod to the logo's hexagon. It
// opens by slicing along a tilted plane and sliding the lid off along it.
package hex

import (
	"math"
	"sort"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/components/vessel"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Hex is the prism.
type Hex struct {
	edges [][2]r.Vec3
	n, p0 r.Vec3 // cut plane
	slide r.Vec3 // direction the lid moves (in plane, upwards)
	hatch r.Vec3 // hatch direction (in plane, horizontal)
	cut   []r.Vec3
	halfH float64 // distance from the center down to the floor

	radius, height float64
}

// Option configures a [Hex].
type Option func(*Hex)

// WithRadius sets the prism's circumradius (default 1.2).
func WithRadius(rad float64) Option { return func(h *Hex) { h.radius = rad } }

// WithHalfHeight sets the distance from the center to the lid and to the
// floor (default 1.05).
func WithHalfHeight(hh float64) Option { return func(h *Hex) { h.height = hh } }

// New returns a hex prism vessel.
func New(opts ...Option) *Hex {
	h := &Hex{radius: 1.2, height: 1.05}
	for _, o := range opts {
		o(h)
	}
	h.edges = prism(h.radius, h.height)
	for _, e := range h.edges {
		h.halfH = math.Max(h.halfH, math.Max(-e[0].Y, -e[1].Y))
	}
	up := r.V(0, 1, 0)
	h.n = r.V(0.55, 1, 0.3).Norm()
	h.p0 = r.V(0, 0.12, 0)
	h.slide = up.Sub(h.n.Scale(h.n.Y)).Norm()
	h.hatch = up.Cross(h.n).Norm()

	// cut polygon: plane ∩ edges, sorted around its centroid
	for _, e := range h.edges {
		da, db := h.dist(e[0]), h.dist(e[1])
		if (da < 0) != (db < 0) {
			h.cut = append(h.cut, e[0].Lerp(e[1], da/(da-db)))
		}
	}
	c := centroid(h.cut)
	sort.Slice(h.cut, func(i, j int) bool {
		ai := math.Atan2(h.cut[i].Sub(c).Dot(h.slide), h.cut[i].Sub(c).Dot(h.hatch))
		aj := math.Atan2(h.cut[j].Sub(c).Dot(h.slide), h.cut[j].Sub(c).Dot(h.hatch))
		return ai < aj
	})
	return h
}

// Name implements [vessel.Vessel].
func (*Hex) Name() string { return "hex" }

// Floor implements [vessel.Vessel].
func (h *Hex) Floor() float64 { return h.halfH }

// Draw implements [vessel.Vessel]: Unseal draws the cut and its hatching,
// Open slides the lid off.
func (h *Hex) Draw(st *stage.Stage, p vessel.Progress) {
	// the prism's front faces sit right over the specimen, so even lines in
	// front of it are dimmed somewhat (and the dense hatching fully)
	defer st.Canvas.Occluding(0.18, 0.35)()
	rot := st.Turntable
	off := h.slide.Scale(0.8 * p.Open).Add(h.n.Scale(0.2 * p.Open))
	vp := st.Pen(rot, st.HeightColor, 0.7)
	up := st.Pen(func(v r.Vec3) r.Vec3 { return rot(v.Add(off)) }, st.HeightColor, 0.7)

	lower, upper := h.split()
	for _, e := range lower {
		vp.Seg(e[0], e[1], p.Build)
	}
	for _, e := range upper {
		up.Seg(e[0], e[1], p.Build)
	}
	if p.Unseal <= 0 {
		return
	}
	cut := append(append([]r.Vec3{}, h.cut...), h.cut[0])
	vp.I = 1.15
	vp.Poly(cut, p.Unseal)
	up.I = 1.15
	if p.Open > 0.02 {
		up.Poly(cut, 1)
	}
	vp.I, up.I = 0.5, 0.5
	st.Canvas.OccludeFront = 0.18
	for k, hl := range h.hatches(9) {
		f := anim.Clamp(p.Unseal*1.6-float64(k)*0.07, 0, 1)
		vp.Seg(hl[0], hl[1], f)
		if p.Open > 0.05 {
			up.Seg(hl[0], hl[1], f)
		}
	}
}

func prism(rad, hh float64) [][2]r.Vec3 {
	var top, bot [6]r.Vec3
	for i := range 6 {
		a := float64(i)*math.Pi/3 + math.Pi/6
		s, c := math.Sincos(a)
		top[i] = r.V(c*rad, hh, s*rad)
		bot[i] = r.V(c*rad, -hh, s*rad)
	}
	edges := make([][2]r.Vec3, 0, 18)
	for i := range 6 {
		j := (i + 1) % 6
		edges = append(edges, [2]r.Vec3{top[i], top[j]}, [2]r.Vec3{bot[i], bot[j]}, [2]r.Vec3{bot[i], top[i]})
	}
	return edges
}

func (h *Hex) dist(p r.Vec3) float64 { return h.n.Dot(p.Sub(h.p0)) }

// split returns the parts of every edge below and above the cut plane.
func (h *Hex) split() (lower, upper [][2]r.Vec3) {
	for _, e := range h.edges {
		da, db := h.dist(e[0]), h.dist(e[1])
		switch {
		case da <= 0 && db <= 0:
			lower = append(lower, e)
		case da > 0 && db > 0:
			upper = append(upper, e)
		default:
			m := e[0].Lerp(e[1], da/(da-db))
			if da <= 0 {
				lower = append(lower, [2]r.Vec3{e[0], m})
				upper = append(upper, [2]r.Vec3{m, e[1]})
			} else {
				upper = append(upper, [2]r.Vec3{e[0], m})
				lower = append(lower, [2]r.Vec3{m, e[1]})
			}
		}
	}
	return lower, upper
}

// hatches returns n evenly spaced segments filling the cut polygon.
func (h *Hex) hatches(n int) [][2]r.Vec3 {
	c := centroid(h.cut)
	type p2 struct{ a, b float64 }
	poly := make([]p2, len(h.cut))
	bmin, bmax := math.Inf(1), math.Inf(-1)
	for i, p := range h.cut {
		d := p.Sub(c)
		poly[i] = p2{d.Dot(h.hatch), d.Dot(h.slide)}
		bmin, bmax = math.Min(bmin, poly[i].b), math.Max(bmax, poly[i].b)
	}
	var out [][2]r.Vec3
	for k := 1; k <= n; k++ {
		b := bmin + (bmax-bmin)*float64(k)/float64(n+1)
		lo, hi := math.Inf(1), math.Inf(-1)
		for i := range poly {
			p, q := poly[i], poly[(i+1)%len(poly)]
			if (p.b-b)*(q.b-b) > 0 || p.b == q.b {
				continue
			}
			a := p.a + (q.a-p.a)*(b-p.b)/(q.b-p.b)
			lo, hi = math.Min(lo, a), math.Max(hi, a)
		}
		if lo < hi {
			base := c.Add(h.slide.Scale(b))
			out = append(out, [2]r.Vec3{base.Add(h.hatch.Scale(lo)), base.Add(h.hatch.Scale(hi))})
		}
	}
	return out
}

func centroid(pts []r.Vec3) r.Vec3 {
	var c r.Vec3
	for _, p := range pts {
		c = c.Add(p)
	}
	return c.Scale(1 / float64(len(pts)))
}
