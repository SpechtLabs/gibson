// Package dodeca is a dodecahedron vessel resting on a face. It opens as an
// exploded view: the faces push out along their normals, the top ones
// furthest, shrinking and twisting a little so clean gaps open up.
package dodeca

import (
	"math"
	"sort"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/components/vessel"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Dodeca is the dodecahedron.
type Dodeca struct {
	panels []panel
	halfH  float64 // distance from the center down to the floor
	scale  float64
}

type panel struct {
	pts    []r.Vec3 // closed loop
	center r.Vec3
	normal r.Vec3
}

// Option configures a [Dodeca].
type Option func(*Dodeca)

// WithScale scales the solid from edge length 2/φ (default 0.78, a
// circumradius of about 1.35).
func WithScale(s float64) Option { return func(d *Dodeca) { d.scale = s } }

// New returns a dodecahedron vessel.
func New(opts ...Option) *Dodeca {
	d := &Dodeca{scale: 0.78}
	for _, o := range opts {
		o(d)
	}
	var edges [][2]r.Vec3
	edges, d.panels = solid(d.scale)
	for _, e := range edges {
		d.halfH = math.Max(d.halfH, math.Max(-e[0].Y, -e[1].Y))
	}
	return d
}

// Name implements [vessel.Vessel].
func (*Dodeca) Name() string { return "dodeca" }

// Floor implements [vessel.Vessel].
func (d *Dodeca) Floor() float64 { return d.halfH }

// Draw implements [vessel.Vessel]: Unseal flares the seams, Open explodes
// the faces outwards, and while Idle they breathe.
func (d *Dodeca) Draw(st *stage.Stage, p vessel.Progress) {
	t := st.T
	up := r.V(0, 1, 0)
	for i, pn := range d.panels {
		lift := (pn.normal.Y + 1) / 2 // 0 for the floor, 1 for the lid
		fi := float64(i)
		off := pn.normal.Scale(0.8 * p.Open * lift).Add(up.Scale(0.45 * p.Open * lift * lift))
		if p.Idle {
			off = off.Add(pn.normal.Scale(0.04 * lift * math.Sin(t*1.3+fi)))
		}
		shrink := 1 - 0.3*p.Open*lift
		twist := 0.5 * p.Open * lift
		if i%2 == 1 {
			twist = -twist
		}
		xf := func(v r.Vec3) r.Vec3 {
			dv := v.Sub(pn.center).RotAxis(pn.normal, twist).Scale(shrink)
			return st.Turntable(pn.center.Add(off).Add(dv))
		}
		// seams flare while the panels unlatch; shared edges overlap while
		// closed, so each face draws at roughly half strength then
		flare := 1 + 0.8*math.Sin(p.Unseal*math.Pi)
		st.Pen(xf, st.HeightColor, 0.4*flare*(1+0.6*anim.Clamp(p.Open*4, 0, 1))).Poly(pn.pts, p.Build)
	}
}

func solid(scale float64) ([][2]r.Vec3, []panel) {
	phi := (1 + math.Sqrt(5)) / 2
	ip := 1 / phi
	pts := make([]r.Vec3, 0, 20)
	for _, a := range []float64{-1, 1} {
		for _, b := range []float64{-1, 1} {
			for _, c := range []float64{-1, 1} {
				pts = append(pts, r.V(a, b, c))
			}
			pts = append(pts, r.V(0, a*ip, b*phi), r.V(a*ip, b*phi, 0), r.V(a*phi, 0, b*ip))
		}
	}
	// rotate so the face with normal (0,φ,1) points up (and its opposite down)
	tilt := -math.Atan(ip)
	for i := range pts {
		pts[i] = pts[i].RotX(tilt).Scale(scale)
	}
	edgeLen := 2 * ip * scale
	edges := make([][2]r.Vec3, 0, 30)
	for i := range pts {
		for j := i + 1; j < len(pts); j++ {
			if math.Abs(pts[i].Sub(pts[j]).Len()-edgeLen) < 1e-6 {
				edges = append(edges, [2]r.Vec3{pts[i], pts[j]})
			}
		}
	}
	// faces: one per icosahedron vertex direction, the 5 closest vertices
	panels := make([]panel, 0, 12)
	for _, a := range []float64{-1, 1} {
		for _, b := range []float64{-1, 1} {
			for _, n := range []r.Vec3{r.V(0, a*phi, b), r.V(a*phi, b, 0), r.V(a, 0, b*phi)} {
				n = n.Norm().RotX(tilt)
				var face []r.Vec3
				best := 0.0
				for _, p := range pts {
					best = math.Max(best, p.Dot(n))
				}
				for _, p := range pts {
					if p.Dot(n) > best-1e-6 {
						face = append(face, p)
					}
				}
				var c r.Vec3
				for _, p := range face {
					c = c.Add(p)
				}
				c = c.Scale(1 / float64(len(face)))
				u := face[0].Sub(c).Norm()
				w := n.Cross(u)
				sort.Slice(face, func(i, j int) bool {
					return math.Atan2(face[i].Sub(c).Dot(w), face[i].Sub(c).Dot(u)) <
						math.Atan2(face[j].Sub(c).Dot(w), face[j].Sub(c).Dot(u))
				})
				panels = append(panels, panel{pts: append(face, face[0]), center: c, normal: n})
			}
		}
	}
	return edges, panels
}
