// Package bird is the woodpecker specimen, traced from the Specht Labs logo.
package bird

import (
	"math"
	"sort"

	"github.com/spechtlabs/gibson/pkg/components/specimen"
	r "github.com/spechtlabs/gibson/pkg/render"
)

// Woodpecker side profile, traced loosely from the logo (image coordinates,
// 756px square, beak pointing left).
var outline = [][2]float64{
	{97, 100}, {160, 94}, {215, 90}, {235, 72}, {270, 58}, {310, 55}, {345, 65},
	{372, 88}, {385, 120}, {382, 160}, {400, 185}, {440, 215}, {480, 255},
	{515, 305}, {548, 360}, {582, 415}, {615, 465}, {645, 515}, {660, 550},
	{625, 540}, {585, 512}, {600, 600}, {610, 690}, {585, 695}, {555, 620},
	{520, 545}, {480, 505}, {450, 470}, {410, 455}, {360, 450}, {330, 448},
	{295, 462}, {280, 472}, {298, 446}, {270, 420}, {250, 380}, {236, 320},
	{232, 260}, {226, 205}, {214, 162}, {212, 130}, {160, 114}, {97, 100},
}

var wing = [][2]float64{
	{300, 250}, {335, 222}, {390, 215}, {440, 240}, {480, 285}, {510, 335}, {520, 365},
}

var cheek = [][2]float64{{214, 130}, {260, 165}, {320, 178}, {372, 165}}

var feathers = [][2]float64{{430, 380}, {600, 540}}

func pt(p [2]float64) r.Vec3 { return r.V((p[0]-390)/330, -(p[1]-380)/330, 0) }

// Bird is the woodpecker. Its model doesn't move, so it's built once.
type Bird struct {
	strands []specimen.Strand
}

// New returns the woodpecker.
func New() *Bird { return &Bird{strands: model()} }

// Name implements [specimen.Specimen].
func (*Bird) Name() string { return "bird" }

// Designation implements [specimen.Specimen].
func (*Bird) Designation() string { return "PICIDAE // DENDROCOPOS MAJOR" }

// Scale implements [specimen.Specimen].
func (*Bird) Scale() float64 { return 1.12 }

// Strands implements [specimen.Specimen].
func (b *Bird) Strands(float64) []specimen.Strand { return b.strands }

// model builds a hologram-style woodpecker: a bright profile outline, a
// faint back contour for depth, horizontal scanline hatching filling the
// silhouette and a few detail lines.
func model() []specimen.Strand {
	var out []specimen.Strand
	line := make([]r.Vec3, len(outline))
	for i, p := range outline {
		line[i] = pt(p)
	}
	at := func(pts []r.Vec3, z float64) []r.Vec3 {
		o := make([]r.Vec3, len(pts))
		for i, v := range pts {
			o[i] = r.V(v.X, v.Y, z)
		}
		return o
	}
	out = append(out, specimen.Strand{Pts: line, I: 1.1}, specimen.Strand{Pts: at(line, 0.09), I: 0.3}, specimen.Strand{Pts: at(line, -0.09), I: 0.3})

	// scanline fill (even-odd over the non-convex outline)
	for y := -0.93; y < 0.97; y += 0.075 {
		var xs []float64
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			if (a.Y-y)*(b.Y-y) >= 0 {
				continue
			}
			xs = append(xs, a.X+(b.X-a.X)*(y-a.Y)/(b.Y-a.Y))
		}
		sort.Float64s(xs)
		for k := 0; k+1 < len(xs); k += 2 {
			out = append(out, specimen.Strand{Pts: []r.Vec3{r.V(xs[k]+0.015, y, 0), r.V(xs[k+1]-0.015, y, 0)}, I: 0.32})
		}
	}

	for _, detail := range [][][2]float64{wing, cheek, feathers} {
		pl := make([]r.Vec3, 0, len(detail))
		for _, p := range detail {
			pl = append(pl, pt(p))
		}
		out = append(out, specimen.Strand{Pts: pl, I: 0.9})
	}
	e := pt([2]float64{262, 108})
	eye := make([]r.Vec3, 0, 13)
	for k := 0; k <= 12; k++ {
		s, c := math.Sincos(float64(k) / 12 * 2 * math.Pi)
		eye = append(eye, r.V(e.X+c*0.04, e.Y+s*0.04, 0))
	}
	return append(out, specimen.Strand{Pts: eye, I: 1.4, Accent: true})
}
