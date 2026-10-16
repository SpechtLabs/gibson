// Package containment defines the containment field that holds the
// specimen once it has materialized. Each field lives in a sub-package of
// its own and returns a [Field].
package containment

import "github.com/spechtlabs/gibson/pkg/stage"

// Field is a containment field.
type Field interface {
	// Name is what it's picked by.
	Name() string
	// Draw draws the field, in from 0 (off) to 1 (fully up).
	Draw(st *stage.Stage, in float64)
}
