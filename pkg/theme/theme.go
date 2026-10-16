// Package theme defines the palette every part of the scene draws with. Each
// theme lives in a sub-package of its own and returns a [Theme].
package theme

import "github.com/spechtlabs/gibson/pkg/render"

// Theme is a named palette.
type Theme interface {
	Name() string
	Palette() Palette
}

// Palette assigns a color to every role in the scene.
type Palette struct {
	Grad     []render.RGB // wireframe color by height, top -> bottom
	Specimen render.RGB   // the hot thing in the middle
	Eye      render.RGB   // specimen accent (eye)
	Rings    render.RGB   // containment field
	Orbit    render.RGB   // projector floor rings
	Beam     render.RGB   // projector beam / accent
	HUD      render.RGB
	Text     render.RGB
	Warn     render.RGB
	BG       render.RGB // background at the center of the screen
	Title    []render.RGB
}

// New returns a Theme with a fixed palette, for themes that need nothing
// more.
func New(name string, p Palette) Theme { return static{name: name, p: p} }

type static struct {
	name string
	p    Palette
}

func (s static) Name() string     { return s.name }
func (s static) Palette() Palette { return s.p }
