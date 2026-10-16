// Package title is the big block-letter title. Its letters decrypt out of
// noise one by one, flash white as they lock in, and a sheen sweeps across
// them while the scene holds. On small terminals it falls back to a single
// line of spaced-out capitals.
package title

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
)

// Title is the title.
type Title struct {
	text  string
	start float64
	sheen float64
}

// Option configures a [Title].
type Option func(*Title)

// WithStart sets when the title starts decrypting (default 4.7s).
func WithStart(t float64) Option { return func(ti *Title) { ti.start = t } }

// WithSheenFrom sets when the sheen starts sweeping (default 7s).
func WithSheenFrom(t float64) Option { return func(ti *Title) { ti.sheen = t } }

// New returns a title reading text, which [font] has to cover.
func New(text string, opts ...Option) *Title {
	ti := &Title{text: text, start: 4.7, sheen: 7}
	for _, o := range opts {
		o(ti)
	}
	return ti
}

// Draw implements [stage.Layer].
func (ti *Title) Draw(st *stage.Stage) {
	if st.T < ti.start {
		return
	}
	if st.Layout.Compact {
		ti.drawCompact(st)
		return
	}
	ti.drawBlocks(st)
}

type px struct {
	on  bool
	col r.RGB
}

func (ti *Title) drawBlocks(st *stage.Stage) {
	t := st.T
	sc := st.Layout.TitleScale
	gw, gap, wordGap := 5*sc, sc*2, sc*3
	// pixel buffer for the title
	pw := 0
	for _, ch := range ti.text {
		if ch == ' ' {
			pw += wordGap
		} else {
			pw += gw + gap
		}
	}
	pw -= gap
	ph := 7 * sc
	buf := make([]px, pw*ph)
	x0 := 0
	q := uint32(t * 18)
	for li, ch := range ti.text {
		if ch == ' ' {
			x0 += wordGap
			continue
		}
		resolve := ti.start + 0.25 + float64(li)*0.09 + anim.Hash(uint32(li), 11, 0)*0.15
		ti.glyph(st, buf, pw, ph, x0, li, font[ch], resolve, q)
		x0 += gw + gap
	}

	// two pixel rows per cell, as half blocks
	ox := (st.Layout.W - pw) / 2
	oy := st.Layout.TitleY
	for cy := 0; cy*2 < ph; cy++ {
		for x := range pw {
			top := buf[(cy*2)*pw+x]
			var bot px
			if cy*2+1 < ph {
				bot = buf[(cy*2+1)*pw+x]
			}
			switch {
			case top.on && bot.on:
				st.Canvas.SetCellBG(ox+x, oy+cy, '▀', top.col, bot.col, 0.55)
			case top.on:
				st.Canvas.SetCell(ox+x, oy+cy, '▀', top.col, 0.55)
			case bot.on:
				st.Canvas.SetCell(ox+x, oy+cy, '▄', bot.col, 0.55)
			}
		}
	}
}

// glyph renders letter li (glyph g) into buf at x0: noise until resolve,
// then the glyph flashing white, then the sheen once the scene holds.
func (ti *Title) glyph(st *stage.Stage, buf []px, pw, ph, x0, li int, g [7]string, resolve float64, q uint32) {
	t := st.T
	sc := st.Layout.TitleScale
	for gy := range ph {
		for gx := range 5 * sc {
			var on bool
			if t >= resolve {
				on = lit(g, gx/sc, gy/sc)
			} else {
				on = anim.Hash(uint32(li*97+gx/sc), uint32(gy/sc), q) < 0.42
			}
			if on {
				buf[gy*pw+x0+gx] = px{true, ti.color(st, pw, ph, x0+gx, gy, resolve)}
			}
		}
	}
}

func (ti *Title) color(st *stage.Stage, pw, ph, x, gy int, resolve float64) r.RGB {
	t := st.T
	white := r.RGB{R: 1, G: 1, B: 1}
	col := r.Ramp(st.Palette.Title, float64(gy)/float64(ph-1))
	if t < resolve {
		col = col.Scale(0.45)
	} else if f := 1 - (t-resolve)/0.35; f > 0 {
		col = col.Lerp(white, f)
	}
	// periodic sheen sweeping across
	if t > ti.sheen {
		sweep := math.Mod((t-ti.sheen)*0.45, 1.6)*float64(pw+20) - 10
		d := float64(x) - sweep + float64(gy)*0.6
		if d > -2 && d < 2 {
			col = col.Lerp(white, 0.6*(1-math.Abs(d)/2))
		}
	}
	return col
}

// drawCompact spaces the title's letters out on one line, scrambling each
// through glyph noise until it resolves.
func (ti *Title) drawCompact(st *stage.Stage) {
	t := st.T
	// one space between letters, so three between words
	runes := make([]rune, 0, 2*len(ti.text))
	for i, ch := range []rune(ti.text) {
		if i > 0 {
			runes = append(runes, ' ')
		}
		runes = append(runes, ch)
	}
	ox := (st.Layout.W - len(runes)) / 2
	glyphs := []rune("!<>-_\\/[]{}=+*^?#░▒")
	for i, ch := range runes {
		if ch == ' ' {
			continue
		}
		resolve := ti.start + 0.2 + float64(i)*0.04
		col := r.Ramp(st.Palette.Title, 0.5)
		if t < resolve {
			ch = glyphs[int(anim.Hash(uint32(i), uint32(t*20), 5)*float64(len(glyphs)))%len(glyphs)]
			col = col.Scale(0.5)
		}
		st.Canvas.SetCell(ox+i, st.Layout.TitleY, ch, col, 0.6)
	}
}
