// Package umbrella is the corporate biohazard theme: white and blood red.
package umbrella

import (
	"github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/theme"
)

var hex = render.Hex

// New returns the umbrella theme.
func New() theme.Theme {
	return theme.New("umbrella", theme.Palette{
		Grad:     []render.RGB{hex(0xffffff), hex(0xff6a5c), hex(0xe0101e), hex(0x7a0a14)},
		Specimen: hex(0xfff1ec),
		Eye:      hex(0xff1a2a),
		Rings:    hex(0xff1a2a),
		Orbit:    hex(0x6a1018),
		Beam:     hex(0xffffff),
		HUD:      hex(0x7a1a20),
		Text:     hex(0xffd6d0),
		Warn:     hex(0xffd23a),
		BG:       hex(0x0d0204),
		Title:    []render.RGB{hex(0xffffff), hex(0xffd0cc), hex(0xff2a36), hex(0x9a0a16)},
	})
}
