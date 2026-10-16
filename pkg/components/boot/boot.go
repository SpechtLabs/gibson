// Package boot is the boot log: lines of a fake BIOS typing out in the
// corner while the scene comes up, the last one blinking as an alert.
package boot

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Log is the boot log.
type Log struct {
	lines        []string
	start, fade  float64
	x, y         int
	charsPerSec  float64
	pausePerLine int
}

// Option configures a [Log].
type Option func(*Log)

// WithLines sets the lines typed out; the last one blinks.
func WithLines(lines ...string) Option { return func(l *Log) { l.lines = lines } }

// WithStart sets when typing starts (default 0.45s).
func WithStart(t float64) Option { return func(l *Log) { l.start = t } }

// WithFadeAt sets when the log fades out, over half a second (default 2.9s).
func WithFadeAt(t float64) Option { return func(l *Log) { l.fade = t } }

// New returns a boot log.
func New(opts ...Option) *Log {
	l := &Log{
		lines:        []string{"BIOS ......................... OK", ">> READY"},
		start:        0.45,
		fade:         2.9,
		x:            3,
		y:            2,
		charsPerSec:  120,
		pausePerLine: 6,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Draw implements [stage.Layer].
func (l *Log) Draw(st *stage.Stage) {
	t := st.T
	pal := st.Palette
	fade := 1 - anim.Phase(t, l.fade, l.fade+0.5)
	if fade <= 0 {
		return
	}
	chars := int((t - l.start) * l.charsPerSec)
	last := len(l.lines) - 1
	for i, line := range l.lines {
		if chars <= 0 {
			break
		}
		n := min(chars, len(line))
		col := pal.Text
		if i == last {
			col = pal.Warn
			if math.Mod(t, 0.4) < 0.15 {
				col = col.Scale(0.4)
			}
		}
		st.Canvas.Text(l.x, l.y+i, line[:n], col.Scale(fade), 0.3)
		if (n < len(line) || i == last) && math.Mod(t, 0.5) < 0.3 {
			st.Canvas.SetCell(l.x+n, l.y+i, '█', pal.Text.Scale(fade), 0.3)
		}
		chars -= len(line) + l.pausePerLine
	}
}
