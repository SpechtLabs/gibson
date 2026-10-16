// Package vessel defines the containment vessel: the wireframe solid around
// the specimen that builds itself, unseals and opens. Each shape lives in a
// sub-package of its own and returns a [Vessel] with its own way of opening.
package vessel

import "github.com/spechtlabs/gibson/pkg/stage"

// Vessel is a containment vessel.
type Vessel interface {
	// Name is what it's picked by, e.g. on the command line.
	Name() string
	// Floor is how far below the center the vessel's floor sits, where the
	// projector stands.
	Floor() float64
	// Draw draws the vessel at progress p, turned by the stage's yaw.
	Draw(st *stage.Stage, p Progress)
}

// Progress is how far a vessel is through its sequence.
type Progress struct {
	Build  float64 // the wireframe drawing in, 0 to 1
	Unseal float64 // the seams or the cut appearing, 0 to 1
	Open   float64 // opening, 0 to 1, overshooting on the way
	Idle   bool    // the sequence is holding; the vessel may drift
}
