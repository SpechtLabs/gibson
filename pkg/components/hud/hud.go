// Package hud is the heads-up display framing the scene: corner brackets, a
// title bar, a clock with a recording dot, the specimen readout, telemetry
// and a vertical warning that blinks while something is unstable.
package hud

import (
	"fmt"
	"math"
	"strings"

	"github.com/spechtlabs/gibson/pkg/anim"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Readout is what the HUD reports about the specimen at a moment.
type Readout struct {
	Designation string // e.g. PICIDAE // DENDROCOPOS MAJOR
	Status      string // e.g. MATERIALIZING
	Settled     bool   // a settled status stays calm; any other blinks
	Warning     bool   // show the vertical WARNING / UNSTABLE banner
}

// HUD is the display.
type HUD struct {
	start   float64
	label   string
	labelAt float64
	readout func(t float64) Readout
	anchor  r.Vec3 // world point the warning banner hangs from
}

// Option configures a [HUD].
type Option func(*HUD)

// WithStart sets when the HUD flickers on (default 1s).
func WithStart(t float64) Option { return func(h *HUD) { h.start = t } }

// WithLabel sets the title bar's text and when it starts typing out (it
// begins 0.4s after at).
func WithLabel(label string, at float64) Option {
	return func(h *HUD) { h.label, h.labelAt = label, at }
}

// WithReadout sets what the HUD reports at time t.
func WithReadout(f func(t float64) Readout) Option { return func(h *HUD) { h.readout = f } }

// New returns a HUD.
func New(opts ...Option) *HUD {
	h := &HUD{
		start:   1,
		readout: func(float64) Readout { return Readout{Status: "NOMINAL", Settled: true} },
		anchor:  r.V(-1.7, 0, 0),
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// Draw implements [stage.Layer].
func (h *HUD) Draw(st *stage.Stage) {
	t := st.T
	if t < h.start {
		return
	}
	// flicker on
	if t < h.start+0.35 && anim.Hash(uint32(t*30), 3, 3) < 0.5 {
		return
	}
	h.corners(st)
	if st.Layout.Compact {
		return
	}
	pal := st.Palette
	cv := st.Canvas
	w, ht := st.Layout.W, st.Layout.H
	txt := pal.Text.Scale(0.75)
	dim := pal.HUD.Lerp(pal.Text, 0.35)

	// top-left title bar
	if h.label != "" && t > h.labelAt+0.4 {
		n := min(len([]rune(h.label)), int((t-h.labelAt-0.4)*60))
		cv.Text(3, 1, string([]rune(h.label)[:n]), txt, 0.2)
	}
	// top-right clock + live dot
	clock := fmt.Sprintf("T+%02d:%05.2f", int(t)/60, math.Mod(t, 60))
	cv.Text(w-3-len(clock)-7, 1, clock, dim, 0.1)
	if math.Mod(t, 1) < 0.6 {
		cv.SetCell(w-3-5, 1, '●', pal.Rings, 0.6)
	}
	cv.Text(w-3-3, 1, "REC", dim, 0.1)

	// bottom-left specimen readout
	ro := h.readout(t)
	stCol := pal.Warn
	if ro.Settled {
		stCol = pal.Text
	} else if math.Mod(t, 0.5) < 0.2 {
		stCol = stCol.Scale(0.5)
	}
	cv.Text(3, ht-3, "SPECIMEN  "+ro.Designation, dim, 0.1)
	cv.Text(3, ht-2, "STATUS    ", dim, 0.1)
	cv.Text(13, ht-2, ro.Status, stCol, 0.4)

	// bottom-right telemetry
	yaw := math.Mod(st.Yaw*180/math.Pi, 360)
	flux := 0.55 + 0.35*math.Sin(t*1.7) + 0.1*anim.Hash(uint32(t*8), 1, 9)
	bar := strings.Repeat("▮", int(flux*10)) + strings.Repeat("▯", 10-int(flux*10))
	l1 := fmt.Sprintf("YAW %6.1f°", yaw)
	l2 := "FLUX " + bar
	cv.Text(w-3-len([]rune(l2)), ht-3, l1, dim, 0.1)
	cv.Text(w-3-len([]rune(l2)), ht-2, l2, dim, 0.1)

	// vertical warning, blinking
	if ro.Warning && math.Mod(t, 0.45) < 0.3 {
		x, _, _ := st.Project(h.anchor)
		cx := int(x/2) - 2
		_, y0, _ := st.Project(h.anchor.Add(r.V(0, 0.75, 0)))
		cy := int(y0 / 4)
		for i, ch := range "WARNING" {
			cv.SetCell(cx, cy+i, ch, pal.Warn, 0.9)
		}
		for i, ch := range "UNSTABLE" {
			cv.SetCell(cx+2, cy+i, ch, pal.Rings.Scale(0.8), 0.4)
		}
	}
}

// corners draws the four corner brackets.
func (h *HUD) corners(st *stage.Stage) {
	c := st.Palette.HUD
	w, ht := st.Layout.W, st.Layout.H
	k := 4
	if w < 40 {
		k = 2
	}
	corner := func(x, y, dx, dy int, cr rune) {
		st.Canvas.SetCell(x, y, cr, c, 0.2)
		for i := 1; i <= k; i++ {
			st.Canvas.SetCell(x+dx*i, y, '─', c, 0.2)
		}
		st.Canvas.SetCell(x, y+dy, '│', c, 0.2)
	}
	corner(1, 0, 1, 1, '┌')
	corner(w-2, 0, -1, 1, '┐')
	corner(1, ht-1, 1, -1, '└')
	corner(w-2, ht-1, -1, -1, '┘')
}
