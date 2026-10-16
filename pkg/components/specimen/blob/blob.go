// Package blob is an amorphous specimen: a sphere whose surface churns.
package blob

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/components/specimen"
	r "github.com/spechtlabs/gibson/pkg/render"
)

// Blob is a noise-deformed lat/long sphere, regenerated every frame.
type Blob struct{}

// New returns the blob.
func New() *Blob { return &Blob{} }

// Name implements [specimen.Specimen].
func (*Blob) Name() string { return "blob" }

// Designation implements [specimen.Specimen].
func (*Blob) Designation() string { return "UNKNOWN // AMORPHOUS" }

// Scale implements [specimen.Specimen].
func (*Blob) Scale() float64 { return 0.9 }

// Strands implements [specimen.Specimen].
func (*Blob) Strands(t float64) []specimen.Strand {
	rad := func(th, ph float64) float64 {
		return 0.62 * (1 + 0.16*math.Sin(3*ph+2.1*t)*math.Sin(2*th+1.3*t) + 0.09*math.Sin(5*th-1.7*t+ph))
	}
	pt := func(th, ph float64) r.Vec3 {
		st, ct := math.Sincos(th)
		sp, cp := math.Sincos(ph)
		k := rad(th, ph)
		return r.V(st*cp*k, ct*k, st*sp*k)
	}
	const lat, lon, seg = 9, 12, 28
	lines := make([]specimen.Strand, 0, lat-1+lon)
	for i := 1; i < lat; i++ {
		th := float64(i) / lat * math.Pi
		pl := make([]r.Vec3, 0, seg+1)
		for k := 0; k <= seg; k++ {
			pl = append(pl, pt(th, float64(k)/seg*2*math.Pi))
		}
		lines = append(lines, specimen.Strand{Pts: pl, I: 0.8})
	}
	for j := range lon {
		ph := float64(j) / lon * 2 * math.Pi
		pl := make([]r.Vec3, 0, seg+1)
		for k := 0; k <= seg; k++ {
			pl = append(pl, pt(float64(k)/seg*math.Pi, ph))
		}
		lines = append(lines, specimen.Strand{Pts: pl, I: 0.8})
	}
	return lines
}
