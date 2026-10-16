package scene

import (
	"math"
	"sort"

	r "github.com/spechtlabs/gibson/internal/render"
)

// ---------------------------------------------------------------- vessel --

// The vessel is a wireframe solid that gets sliced along a tilted plane; the
// upper half slides off along that plane. Shapes: "hex" (a prism, a nod to
// the hex logo) or "dodeca" (a dodecahedron resting on a face).
type vessel struct {
	edges [][2]r.Vec3
	n, p0 r.Vec3 // cut plane
	slide r.Vec3 // direction the lid moves (in plane, upwards)
	hatch r.Vec3 // hatch direction (in plane, horizontal)
	cut   []r.Vec3
	halfH float64 // distance from the center down to the floor

	// panels, if set, switch the opening from slice-and-slide to an exploded
	// view: every face drifts out along its normal.
	panels []panel
}

type panel struct {
	pts    []r.Vec3 // closed loop
	center r.Vec3
	normal r.Vec3
}

func hexPrism() [][2]r.Vec3 {
	const rad, hh = 1.2, 1.05
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

func dodecahedron() ([][2]r.Vec3, []panel) {
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
	const scale = 0.78 // circumradius √3·0.78 ≈ 1.35
	for i := range pts {
		pts[i] = pts[i].RotX(tilt).Scale(scale)
	}
	edgeLen := 2 * ip * scale
	edges := make([][2]r.Vec3, 0, 18)
	for i := range pts {
		for j := i + 1; j < len(pts); j++ {
			if math.Abs(pts[i].Sub(pts[j]).Len()-edgeLen) < 1e-6 {
				edges = append(edges, [2]r.Vec3{pts[i], pts[j]})
			}
		}
	}
	// faces: one per icosahedron vertex direction, the 5 closest vertices
	var panels []panel
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

// rotAxis rotates v around the unit axis k (Rodrigues).
func rotAxis(v, k r.Vec3, a float64) r.Vec3 {
	s, c := math.Sincos(a)
	return v.Scale(c).Add(k.Cross(v).Scale(s)).Add(k.Scale(k.Dot(v) * (1 - c)))
}

func newVessel(shape string) *vessel {
	v := &vessel{}
	switch shape {
	case "dodeca":
		v.edges, v.panels = dodecahedron()
	default:
		v.edges = hexPrism()
	}
	for _, e := range v.edges {
		v.halfH = math.Max(v.halfH, math.Max(-e[0].Y, -e[1].Y))
	}
	up := r.V(0, 1, 0)
	v.n = r.V(0.55, 1, 0.3).Norm()
	v.p0 = r.V(0, 0.12, 0)
	v.slide = up.Sub(v.n.Scale(v.n.Y)).Norm()
	v.hatch = up.Cross(v.n).Norm()

	// cut polygon: plane ∩ edges, sorted around its centroid
	for _, e := range v.edges {
		da, db := v.dist(e[0]), v.dist(e[1])
		if (da < 0) != (db < 0) {
			v.cut = append(v.cut, e[0].Lerp(e[1], da/(da-db)))
		}
	}
	var c r.Vec3
	for _, p := range v.cut {
		c = c.Add(p)
	}
	c = c.Scale(1 / float64(len(v.cut)))
	sort.Slice(v.cut, func(i, j int) bool {
		ai := math.Atan2(v.cut[i].Sub(c).Dot(v.slide), v.cut[i].Sub(c).Dot(v.hatch))
		aj := math.Atan2(v.cut[j].Sub(c).Dot(v.slide), v.cut[j].Sub(c).Dot(v.hatch))
		return ai < aj
	})
	return v
}

func (v *vessel) dist(p r.Vec3) float64 { return v.n.Dot(p.Sub(v.p0)) }

// split returns the parts of every edge below and above the cut plane.
func (v *vessel) split() (lower, upper [][2]r.Vec3) {
	for _, e := range v.edges {
		da, db := v.dist(e[0]), v.dist(e[1])
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
	return
}

// hatches returns n evenly spaced segments filling the cut polygon.
func (v *vessel) hatches(n int) [][2]r.Vec3 {
	var c r.Vec3
	for _, p := range v.cut {
		c = c.Add(p)
	}
	c = c.Scale(1 / float64(len(v.cut)))
	type p2 struct{ a, b float64 }
	poly := make([]p2, len(v.cut))
	bmin, bmax := math.Inf(1), math.Inf(-1)
	for i, p := range v.cut {
		d := p.Sub(c)
		poly[i] = p2{d.Dot(v.hatch), d.Dot(v.slide)}
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
			base := c.Add(v.slide.Scale(b))
			out = append(out, [2]r.Vec3{base.Add(v.hatch.Scale(lo)), base.Add(v.hatch.Scale(hi))})
		}
	}
	return out
}

// ------------------------------------------------------------- specimen --

// Woodpecker side profile, traced loosely from the logo (image coordinates,
// 756px square, beak pointing left).
var birdOutline = [][2]float64{
	{97, 100}, {160, 94}, {215, 90}, {235, 72}, {270, 58}, {310, 55}, {345, 65},
	{372, 88}, {385, 120}, {382, 160}, {400, 185}, {440, 215}, {480, 255},
	{515, 305}, {548, 360}, {582, 415}, {615, 465}, {645, 515}, {660, 550},
	{625, 540}, {585, 512}, {600, 600}, {610, 690}, {585, 695}, {555, 620},
	{520, 545}, {480, 505}, {450, 470}, {410, 455}, {360, 450}, {330, 448},
	{295, 462}, {280, 472}, {298, 446}, {270, 420}, {250, 380}, {236, 320},
	{232, 260}, {226, 205}, {214, 162}, {212, 130}, {160, 114}, {97, 100},
}

var birdWing = [][2]float64{
	{300, 250}, {335, 222}, {390, 215}, {440, 240}, {480, 285}, {510, 335}, {520, 365},
}

var birdCheek = [][2]float64{{214, 130}, {260, 165}, {320, 178}, {372, 165}}

var birdFeathers = [][2]float64{{430, 380}, {600, 540}}

func birdPt(p [2]float64) r.Vec3 { return r.V((p[0]-390)/330, -(p[1]-380)/330, 0) }

type strand struct {
	pts    []r.Vec3
	i      float64
	accent bool
}

// birdModel builds a hologram-style woodpecker: a bright profile outline, a
// faint back contour for depth, horizontal scanline hatching filling the
// silhouette and a few detail lines.
func birdModel() []strand {
	var out []strand
	outline := make([]r.Vec3, len(birdOutline))
	for i, p := range birdOutline {
		outline[i] = birdPt(p)
	}
	at := func(pts []r.Vec3, z float64) []r.Vec3 {
		o := make([]r.Vec3, len(pts))
		for i, v := range pts {
			o[i] = r.V(v.X, v.Y, z)
		}
		return o
	}
	out = append(out, strand{outline, 1.1, false}, strand{at(outline, 0.09), 0.3, false}, strand{at(outline, -0.09), 0.3, false})

	// scanline fill (even-odd over the non-convex outline)
	for y := -0.93; y < 0.97; y += 0.075 {
		var xs []float64
		for i := 0; i+1 < len(outline); i++ {
			a, b := outline[i], outline[i+1]
			if (a.Y-y)*(b.Y-y) >= 0 {
				continue
			}
			xs = append(xs, a.X+(b.X-a.X)*(y-a.Y)/(b.Y-a.Y))
		}
		sort.Float64s(xs)
		for k := 0; k+1 < len(xs); k += 2 {
			out = append(out, strand{[]r.Vec3{r.V(xs[k]+0.015, y, 0), r.V(xs[k+1]-0.015, y, 0)}, 0.32, false})
		}
	}

	for _, detail := range [][][2]float64{birdWing, birdCheek, birdFeathers} {
		var pl []r.Vec3
		for _, p := range detail {
			pl = append(pl, birdPt(p))
		}
		out = append(out, strand{pl, 0.9, false})
	}
	e := birdPt([2]float64{262, 108})
	var eye []r.Vec3
	for k := 0; k <= 12; k++ {
		s, c := math.Sincos(float64(k) / 12 * 2 * math.Pi)
		eye = append(eye, r.V(e.X+c*0.04, e.Y+s*0.04, 0))
	}
	out = append(out, strand{eye, 1.4, true})
	return out
}

// blobModel is a noise-deformed lat/long sphere, regenerated per frame.
func blobModel(t float64) []strand {
	rad := func(th, ph float64) float64 {
		return 0.62 * (1 + 0.16*math.Sin(3*ph+2.1*t)*math.Sin(2*th+1.3*t) + 0.09*math.Sin(5*th-1.7*t+ph))
	}
	pt := func(th, ph float64) r.Vec3 {
		st, ct := math.Sincos(th)
		sp, cp := math.Sincos(ph)
		k := rad(th, ph)
		return r.V(st*cp*k, ct*k, st*sp*k)
	}
	var lines []strand
	const lat, lon, seg = 9, 12, 28
	for i := 1; i < lat; i++ {
		th := float64(i) / lat * math.Pi
		var pl []r.Vec3
		for k := 0; k <= seg; k++ {
			pl = append(pl, pt(th, float64(k)/seg*2*math.Pi))
		}
		lines = append(lines, strand{pl, 0.8, false})
	}
	for j := range lon {
		ph := float64(j) / lon * 2 * math.Pi
		var pl []r.Vec3
		for k := 0; k <= seg; k++ {
			pl = append(pl, pt(float64(k)/seg*math.Pi, ph))
		}
		lines = append(lines, strand{pl, 0.8, false})
	}
	return lines
}
