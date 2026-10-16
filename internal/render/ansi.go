package render

import (
	"strconv"
	"strings"
)

// ANSI serializes a frame using 24-bit color, only emitting SGR sequences
// when a color actually changes.
func (f *Frame) ANSI() string {
	var b strings.Builder
	b.Grow(f.W * f.H * 12)
	for y := 0; y < f.H; y++ {
		var lf, lb [3]uint8
		first := true
		for x := 0; x < f.W; x++ {
			c := f.Cells[y*f.W+x]
			fr, fg, fb := c.FG.Bytes()
			br, bgg, bb := c.BG.Bytes()
			nf, nb := [3]uint8{fr, fg, fb}, [3]uint8{br, bgg, bb}
			if first || nb != lb {
				sgr(&b, 48, nb)
				lb = nb
			}
			if c.R != ' ' && (first || nf != lf) {
				sgr(&b, 38, nf)
				lf = nf
			}
			first = false
			b.WriteRune(c.R)
		}
		b.WriteString("\x1b[0m")
		if y < f.H-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func sgr(b *strings.Builder, kind int, c [3]uint8) {
	b.WriteString("\x1b[")
	b.WriteString(strconv.Itoa(kind))
	b.WriteString(";2;")
	b.WriteString(strconv.Itoa(int(c[0])))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(int(c[1])))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(int(c[2])))
	b.WriteByte('m')
}
