// Package specimen defines what floats in the containment vessel. Each
// specimen lives in a sub-package of its own and returns a [Specimen]; the
// projector materializes any of them the same way.
package specimen

import "github.com/spechtlabs/gibson/pkg/render"

// Specimen is a wireframe model made of strands.
type Specimen interface {
	// Name is what it's picked by, e.g. on the command line.
	Name() string
	// Designation is the HUD's readout for it.
	Designation() string
	// Scale is how large the projector shows it.
	Scale() float64
	// Strands returns the model at time t, in model space: roughly a unit
	// cube around the origin, facing +Z.
	Strands(t float64) []Strand
}

// Strand is one polyline of a model.
type Strand struct {
	Pts    []render.Vec3
	I      float64 // intensity
	Accent bool    // drawn in the theme's accent color (an eye, say)
}
