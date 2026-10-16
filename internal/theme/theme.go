package theme

import (
	"fmt"
	"sort"

	"github.com/spechtlabs/gibson/internal/render"
)

// Theme is a palette for the whole scene.
type Theme struct {
	Name     string
	Grad     []render.RGB // wireframe color by height, top -> bottom
	Specimen render.RGB   // the hot thing in the middle
	Eye      render.RGB   // specimen accent (eye)
	Rings    render.RGB   // containment rings
	Orbit    render.RGB   // projector orbit rings
	Beam     render.RGB   // projector beam / accent
	HUD      render.RGB
	Text     render.RGB
	Warn     render.RGB
	BG       render.RGB // background at the center of the screen
	Title    []render.RGB
}

var hex = render.Hex

var themes = map[string]Theme{
	"synth": {
		Name:     "synth",
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
	},
	"specht": {
		Name:     "specht",
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
	},
	"umbrella": {
		Name:     "umbrella",
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
	},
}

// Get returns the theme called name.
func Get(name string) (Theme, error) {
	t, ok := themes[name]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme %q (have: %v)", name, Names())
	}
	return t, nil
}

// Names lists the themes, sorted.
func Names() []string {
	n := make([]string, 0, len(themes))
	for k := range themes {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}
