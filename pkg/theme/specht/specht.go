// Package specht is the Specht Labs theme: the logo's teal-to-blue gradient
// with a pink specimen and amber accents.
package specht

import (
	"github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/theme"
)

var hex = render.Hex

// New returns the specht theme.
func New() theme.Theme {
	return theme.New("specht", theme.Palette{
		Grad:     []render.RGB{hex(0x5cfff0), hex(0x00c2cf), hex(0x1a7dff), hex(0x3448ff)},
		Specimen: hex(0xff5ccb),
		Eye:      hex(0xffc23a),
		Rings:    hex(0x00f0ff),
		Orbit:    hex(0x2456c8),
		Beam:     hex(0xffc23a),
		HUD:      hex(0x1d5f7a),
		Text:     hex(0xbdf6ff),
		Warn:     hex(0xffb020),
		BG:       hex(0x020b12),
		Title:    []render.RGB{hex(0xffffff), hex(0x8ff7ff), hex(0x00b8d4), hex(0x2a5cff)},
	})
}
