package render

import "math"

// Canvas is a braille sub-pixel canvas: every terminal cell holds a 2x4 grid
// of dots. Lines accumulate light per dot; Compose turns that into cells,
// adds a cheap bloom by bleeding light into neighboring cell backgrounds and
// layers text overlays on top.
type Canvas struct {
	W, H   int // cells
	PW, PH int // sub-pixels

	inten []float64
	wsum  []float64 // sum of squared intensities (color weights)
	col   []RGB

	over  []overlay
	bgAdd []RGB

	// MarkMask records the nearest depth drawn into each cell; Occlude (if
	// >0) scales the intensity of later plots that land *behind* that depth.
	// A per-cell depth test, enough to tuck things behind the specimen.
	depth    []float64
	MarkMask bool
	Occlude  float64
	// OccludeFront (if >0) also dims plots *in front of* a marked depth, for
	// shapes whose front lines would otherwise clutter what is behind them.
	OccludeFront float64
}

type overlay struct {
	set   bool
	r     rune
	fg    RGB
	bg    RGB
	hasBG bool
	emit  float64
}

// NewCanvas returns a canvas of w x h cells.
func NewCanvas(w, h int) *Canvas {
	c := &Canvas{}
	c.Resize(w, h)
	return c
}

func (c *Canvas) Resize(w, h int) {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	if w == c.W && h == c.H {
		c.Clear()
		return
	}
	c.W, c.H, c.PW, c.PH = w, h, w*2, h*4
	c.inten = make([]float64, c.PW*c.PH)
	c.col = make([]RGB, c.PW*c.PH)
	c.wsum = make([]float64, c.PW*c.PH)
	c.over = make([]overlay, w*h)
	c.bgAdd = make([]RGB, w*h)
	c.depth = make([]float64, w*h)
	c.clearDepth()
}

func (c *Canvas) Clear() {
	clear(c.inten)
	clear(c.col)
	clear(c.wsum)
	clear(c.over)
	clear(c.bgAdd)
	c.clearDepth()
}

// Occluding sets Occlude and OccludeFront, and returns a func that puts
// back what they were: defer c.Occluding(0.18, 0.35)().
func (c *Canvas) Occluding(behind, front float64) (restore func()) {
	pb, pf := c.Occlude, c.OccludeFront
	c.Occlude, c.OccludeFront = behind, front
	return func() { c.Occlude, c.OccludeFront = pb, pf }
}

// DilateDepth grows every marked region by one cell so sparse shapes (like
// hatched fills) occlude as a solid silhouette.
func (c *Canvas) DilateDepth() {
	src := append([]float64(nil), c.depth...)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			d := src[y*c.W+x]
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					xx, yy := x+dx, y+dy
					if xx >= 0 && yy >= 0 && xx < c.W && yy < c.H {
						d = math.Min(d, src[yy*c.W+xx])
					}
				}
			}
			c.depth[y*c.W+x] = d
		}
	}
}

func (c *Canvas) clearDepth() {
	for i := range c.depth {
		c.depth[i] = math.Inf(1)
	}
}

// Plot adds light to a single sub-pixel at camera depth z.
func (c *Canvas) Plot(x, y int, z float64, col RGB, i float64) {
	if x < 0 || y < 0 || x >= c.PW || y >= c.PH || i <= 0 {
		return
	}
	if c.MarkMask || c.Occlude > 0 {
		m := (y/4)*c.W + x/2
		switch {
		case c.MarkMask:
			c.depth[m] = math.Min(c.depth[m], z)
		case z > c.depth[m]+0.05:
			i *= c.Occlude
		case c.OccludeFront > 0 && !math.IsInf(c.depth[m], 1):
			i *= c.OccludeFront
		}
	}
	k := y*c.PW + x
	// color is weighted by i² so the brightest contributor dominates a cell
	c.inten[k] += i
	c.wsum[k] += i * i
	c.col[k] = c.col[k].Add(col.Scale(i * i))
}

// Line draws a sub-pixel line with interpolated depth, color and intensity.
func (c *Canvas) Line(x0, y0, z0, x1, y1, z1 float64, c0, c1 RGB, i0, i1 float64) {
	// trivially reject lines fully off-canvas
	if (x0 < 0 && x1 < 0) || (y0 < 0 && y1 < 0) ||
		(x0 >= float64(c.PW) && x1 >= float64(c.PW)) || (y0 >= float64(c.PH) && y1 >= float64(c.PH)) {
		return
	}
	dx, dy := x1-x0, y1-y0
	n := min(int(math.Ceil(math.Max(math.Abs(dx), math.Abs(dy)))), 4*(c.PW+c.PH))
	if n == 0 {
		c.Plot(int(math.Floor(x0)), int(math.Floor(y0)), z0, c0, i0)
		return
	}
	for s := 0; s <= n; s++ {
		t := float64(s) / float64(n)
		c.Plot(int(math.Floor(x0+dx*t)), int(math.Floor(y0+dy*t)), z0+(z1-z0)*t, c0.Lerp(c1, t), i0+(i1-i0)*t)
	}
}

// SetCell places a glyph in a cell, overriding any braille dots there.
// emit makes the glyph contribute to the bloom.
func (c *Canvas) SetCell(x, y int, r rune, fg RGB, emit float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	o := &c.over[y*c.W+x]
	o.set, o.r, o.fg, o.emit = true, r, fg, emit
}

// SetCellBG is SetCell with an explicit background (used for half-block art).
func (c *Canvas) SetCellBG(x, y int, r rune, fg, bg RGB, emit float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	o := &c.over[y*c.W+x]
	o.set, o.r, o.fg, o.bg, o.hasBG, o.emit = true, r, fg, bg, true, emit
}

// Text writes a string starting at cell (x,y). Spaces are transparent.
func (c *Canvas) Text(x, y int, s string, fg RGB, emit float64) {
	i := 0
	for _, r := range s {
		if r != ' ' {
			c.SetCell(x+i, y, r, fg, emit)
		}
		i++
	}
}

// AddBG adds light to a cell's background (flashes, scanlines).
func (c *Canvas) AddBG(x, y int, col RGB) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	c.bgAdd[y*c.W+x] = c.bgAdd[y*c.W+x].Add(col)
}

// braille dot bit for sub-pixel (dx,dy) inside a cell
var brailleBit = [4][2]rune{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}

// Cell is one resolved terminal cell.
type Cell struct {
	R      rune
	FG, BG RGB
}

// Frame is a grid of resolved cells, ready to print.
type Frame struct {
	W, H  int
	Cells []Cell
}

func (f *Frame) At(x, y int) *Cell { return &f.Cells[y*f.W+x] }

// Compose resolves the canvas into a frame. base gives the background color
// per cell (vignette etc.), bloom scales how much light bleeds into
// neighboring backgrounds.
func (c *Canvas) Compose(base func(x, y int) RGB, bloom float64, dst *Frame) *Frame {
	if dst == nil || dst.W != c.W || dst.H != c.H {
		dst = &Frame{W: c.W, H: c.H, Cells: make([]Cell, c.W*c.H)}
	}
	glow := make([]RGB, c.W*c.H)

	for cy := 0; cy < c.H; cy++ {
		for cx := 0; cx < c.W; cx++ {
			var bits rune
			var sum, maxI, wsum float64
			var acc RGB
			for dy := range 4 {
				row := (cy*4 + dy) * c.PW
				for dx := range 2 {
					k := row + cx*2 + dx
					i := c.inten[k]
					if i <= 0 {
						continue
					}
					sum += i
					wsum += c.wsum[k]
					acc = acc.Add(c.col[k])
					if i > maxI {
						maxI = i
					}
					if i > 0.18 {
						bits |= brailleBit[dy][dx]
					}
				}
			}
			cell := &dst.Cells[cy*c.W+cx]
			cell.R = ' '
			cell.FG = RGB{}
			k := cy*c.W + cx
			if sum > 0 {
				avg := acc.Scale(1 / wsum)
				if bits != 0 {
					cell.R = 0x2800 + bits
					b := 0.45 + 0.55*math.Min(maxI, 1.4)
					fg := avg.Scale(b)
					if sum > 4 { // overlapping light runs white-hot
						fg = fg.Lerp(RGB{1, 1, 1}, math.Min((sum-4)/8, 0.3))
					}
					cell.FG = fg
				}
				glow[k] = avg.Scale(math.Min(sum/4, 1))
			}
			if o := c.over[k]; o.set {
				glow[k] = glow[k].Add(o.fg.Scale(o.emit))
			}
		}
	}

	blurred := blur(glow, c.W, c.H)

	for cy := 0; cy < c.H; cy++ {
		for cx := 0; cx < c.W; cx++ {
			k := cy*c.W + cx
			cell := &dst.Cells[k]
			bg := base(cx, cy).Add(blurred[k].Scale(bloom)).Add(c.bgAdd[k])
			cell.BG = bg
			if o := c.over[k]; o.set {
				cell.R = o.r
				cell.FG = o.fg
				if o.hasBG {
					cell.BG = o.bg
				}
			}
		}
	}
	return dst
}

// blur is a separable blur, wider horizontally since cells are ~1:2.
func blur(src []RGB, w, h int) []RGB {
	hw := []float64{1, 0.75, 0.45, 0.22, 0.08}
	vw := []float64{1, 0.5, 0.15}
	tmp := make([]RGB, len(src))
	out := make([]RGB, len(src))
	for y := range h {
		for x := range w {
			var acc RGB
			for d := -4; d <= 4; d++ {
				xx := x + d
				if xx < 0 || xx >= w {
					continue
				}
				acc = acc.Add(src[y*w+xx].Scale(hw[abs(d)]))
			}
			tmp[y*w+x] = acc.Scale(1 / 2.0)
		}
	}
	for y := range h {
		for x := range w {
			var acc RGB
			for d := -2; d <= 2; d++ {
				yy := y + d
				if yy < 0 || yy >= h {
					continue
				}
				acc = acc.Add(tmp[yy*w+x].Scale(vw[abs(d)]))
			}
			out[y*w+x] = acc.Scale(1 / 1.3)
		}
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
