// Package synth is the synthwave theme: hot pink to violet wireframes with
// acid-lime accents.
package synth

import (
	"github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/theme"
)

var hex = render.Hex

// New returns the synth theme.
func New() theme.Theme {
	return theme.New("synth", theme.Palette{
		Grad:     []render.RGB{hex(0xff2e5f), hex(0xff2bb8), hex(0xa63cff), hex(0x5b4dff)},
		Specimen: hex(0xff9cf0),
		Eye:      hex(0xd4ff3a),
		Rings:    hex(0xff3b2f),
		Orbit:    hex(0x6b3cff),
		Beam:     hex(0xd4ff3a),
		HUD:      hex(0x5a3a9a),
		Text:     hex(0xd9c8ff),
		Warn:     hex(0xd4ff3a),
		BG:       hex(0x0b0414),
		Title:    []render.RGB{hex(0xf4fbff), hex(0x7ff3ff), hex(0xff4fd8), hex(0xff2e5f)},
	})
}
