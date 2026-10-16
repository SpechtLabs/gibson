// Package crt is the old monitor the scene plays on: a bright band rolling
// down the screen and a vignette with interlaced rows. The tube powering on
// and the tearing are in poweron and glitch.
package crt

import (
	"math"

	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Scanline is a faint bright band rolling down the screen.
type Scanline struct{}

// Draw implements [stage.Layer].
func (Scanline) Draw(st *stage.Stage) {
	h := st.Layout.H
	y := int(math.Mod(st.T*0.18, 1.3) * float64(h))
	for d := -1; d <= 1; d++ {
		v := 0.035 * (1 - math.Abs(float64(d))*0.5)
		for x := 0; x < st.Layout.W; x++ {
			st.Canvas.AddBG(x, y+d, st.Palette.Grad[2].Scale(v))
		}
	}
}

// Vignette is the background: the theme's color, darker towards the edges,
// every other row dimmed like interlacing. Pass it to [r.Canvas.Compose].
func Vignette(st *stage.Stage) func(x, y int) r.RGB {
	bg := st.Palette.BG
	w, h := st.Layout.W, st.Layout.H
	return func(x, y int) r.RGB {
		dx := (float64(x)/float64(w) - 0.5) * 2
		dy := (float64(y)/float64(h) - 0.5) * 2
		v := 1 - 0.7*math.Min(1, (dx*dx+dy*dy)*0.55)
		if y%2 == 1 {
			v *= 0.82
		}
		return bg.Scale(v)
	}
}
