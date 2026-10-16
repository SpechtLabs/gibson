package scene

import (
	"fmt"
	"math"
	"strings"

	r "github.com/spechtlabs/gibson/internal/render"
)

// ------------------------------------------------------------------ boot --

var bootLog = []string{
	"SPECHT LABS SYS/9000 :: BIOS 9.5.1",
	"MEM 65536K .................. OK",
	"NEURAL LATTICE .............. OK",
	"MOUNT /dev/specimen0 ........ OK",
	"LOAD containment.ko ......... OK",
	">> SPECIMEN SIGNAL DETECTED",
}

func (s *Scene) drawBoot() {
	t := s.t
	fade := 1 - phase(t, tBootFade, tBootFade+0.5)
	if fade <= 0 {
		return
	}
	chars := int((t - tBootStart) * 120)
	y := 2
	for i, line := range bootLog {
		if chars <= 0 {
			break
		}
		n := min(chars, len(line))
		col := s.th.Text
		if i == len(bootLog)-1 {
			col = s.th.Warn
			if math.Mod(t, 0.4) < 0.15 {
				col = col.Scale(0.4)
			}
		}
		s.cv.Text(3, y+i, line[:n], col.Scale(fade), 0.3)
		if n < len(line) || i == len(bootLog)-1 {
			if math.Mod(t, 0.5) < 0.3 {
				s.cv.SetCell(3+n, y+i, '█', s.th.Text.Scale(fade), 0.3)
			}
		}
		chars -= len(line) + 6 // small pause per line
	}
}

// ------------------------------------------------------------------- HUD --

func (s *Scene) drawHUD() {
	t := s.t
	if t < tHUD {
		return
	}
	w, h := s.lay.w, s.lay.h
	// flicker on
	if t < tHUD+0.35 && hash(uint32(t*30), 3, 3) < 0.5 {
		return
	}
	c := s.th.HUD
	k := 4
	if w < 40 {
		k = 2
	}
	corner := func(x, y, dx, dy int, cr rune) {
		s.cv.SetCell(x, y, cr, c, 0.2)
		for i := 1; i <= k; i++ {
			s.cv.SetCell(x+dx*i, y, '─', c, 0.2)
		}
		s.cv.SetCell(x, y+dy, '│', c, 0.2)
	}
	corner(1, 0, 1, 1, '┌')
	corner(w-2, 0, -1, 1, '┐')
	corner(1, h-1, 1, -1, '└')
	corner(w-2, h-1, -1, -1, '┘')

	if s.lay.compact {
		return
	}
	txt := s.th.Text.Scale(0.75)
	dim := s.th.HUD.Lerp(s.th.Text, 0.35)

	// top-left title bar, after the boot log is gone
	if t > tBootFade+0.4 {
		label := "SPECHT LABS ▸ CONTAINMENT UNIT 07"
		n := min(len([]rune(label)), int((t-tBootFade-0.4)*60))
		s.cv.Text(3, 1, string([]rune(label)[:n]), txt, 0.2)
	}
	// top-right clock + live dot
	clock := fmt.Sprintf("T+%02d:%05.2f", int(t)/60, math.Mod(t, 60))
	s.cv.Text(w-3-len(clock)-7, 1, clock, dim, 0.1)
	if math.Mod(t, 1) < 0.6 {
		s.cv.SetCell(w-3-5, 1, '●', s.th.Rings, 0.6)
	}
	s.cv.Text(w-3-3, 1, "REC", dim, 0.1)

	// bottom-left specimen readout
	status := "ACQUIRING"
	switch {
	case t > tMatEnd+0.3:
		status = "CONTAINED"
	case t > tMatStart:
		status = "MATERIALIZING"
	}
	stCol := s.th.Warn
	if status == "CONTAINED" {
		stCol = s.th.Text
	} else if math.Mod(t, 0.5) < 0.2 {
		stCol = stCol.Scale(0.5)
	}
	specimen := "PICIDAE // DENDROCOPOS MAJOR"
	if s.opt.Specimen == specimenBlob {
		specimen = "UNKNOWN // AMORPHOUS"
	}
	s.cv.Text(3, h-3, "SPECIMEN  "+specimen, dim, 0.1)
	s.cv.Text(3, h-2, "STATUS    ", dim, 0.1)
	s.cv.Text(13, h-2, status, stCol, 0.4)

	// bottom-right telemetry
	yaw := math.Mod(s.yaw()*180/math.Pi, 360)
	flux := 0.55 + 0.35*math.Sin(t*1.7) + 0.1*hash(uint32(t*8), 1, 9)
	bar := strings.Repeat("▮", int(flux*10)) + strings.Repeat("▯", 10-int(flux*10))
	l1 := fmt.Sprintf("YAW %6.1f°", yaw)
	l2 := "FLUX " + bar
	s.cv.Text(w-3-len([]rune(l2)), h-3, l1, dim, 0.1)
	s.cv.Text(w-3-len([]rune(l2)), h-2, l2, dim, 0.1)

	// vertical warning while materializing
	if t > tMatStart-0.2 && t < tMatEnd+0.6 && math.Mod(t, 0.45) < 0.3 {
		x, _, _ := s.lay.cam.Project(s.view(r.V(-1.7, 0, 0)))
		cx := int(x/2) - 2
		_, y0, _ := s.lay.cam.Project(s.view(r.V(-1.7, 0.75, 0)))
		cy := int(y0 / 4)
		for i, ch := range "WARNING" {
			s.cv.SetCell(cx, cy+i, ch, s.th.Warn, 0.9)
		}
		for i, ch := range "UNSTABLE" {
			s.cv.SetCell(cx+2, cy+i, ch, s.th.Rings.Scale(0.8), 0.4)
		}
	}
}

// ----------------------------------------------------------------- title --

const titleText = "SPECHT LABS"

var font = map[rune][7]string{
	'S': {".####", "#....", "#....", ".###.", "....#", "....#", "####."},
	'P': {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'C': {".####", "#....", "#....", "#....", "#....", "#....", ".####"},
	'H': {"#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'T': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
	'L': {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'A': {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'B': {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
}

func (s *Scene) drawTitle() {
	t := s.t
	if t < tTitleStart {
		return
	}
	if s.lay.compact {
		s.drawTitleCompact()
		return
	}
	sc := s.lay.titleScale
	gw, gap, wordGap := 5*sc, sc*2, sc*3
	// pixel buffer for the title
	pw := 0
	for _, ch := range titleText {
		if ch == ' ' {
			pw += wordGap
		} else {
			pw += gw + gap
		}
	}
	pw -= gap
	ph := 7 * sc
	type px struct {
		on  bool
		col r.RGB
	}
	buf := make([]px, pw*ph)
	x0 := 0
	q := uint32(t * 18)
	for li, ch := range titleText {
		if ch == ' ' {
			x0 += wordGap
			continue
		}
		g := font[ch]
		resolve := tTitleStart + 0.25 + float64(li)*0.09 + hash(uint32(li), 11, 0)*0.15
		for gy := range ph {
			for gx := range gw {
				var on bool
				if t >= resolve {
					on = g[gy/sc][gx/sc] == '#'
				} else {
					on = hash(uint32(li*97+gx/sc), uint32(gy/sc), q) < 0.42
				}
				if !on {
					continue
				}
				col := r.Ramp(s.th.Title, float64(gy)/float64(ph-1))
				if t < resolve {
					col = col.Scale(0.45)
				} else if f := 1 - (t-resolve)/0.35; f > 0 {
					col = col.Lerp(r.RGB{R: 1, G: 1, B: 1}, f)
				}
				// periodic sheen sweeping across
				if t > tHold {
					sweep := math.Mod((t-tHold)*0.45, 1.6)*float64(pw+20) - 10
					d := float64(x0+gx) - sweep + float64(gy)*0.6
					if d > -2 && d < 2 {
						col = col.Lerp(r.RGB{R: 1, G: 1, B: 1}, 0.6*(1-math.Abs(d)/2))
					}
				}
				buf[gy*pw+x0+gx] = px{true, col}
			}
		}
		x0 += gw + gap
	}

	ox := (s.lay.w - pw) / 2
	oy := s.lay.titleY
	for cy := 0; cy*2 < ph; cy++ {
		for x := 0; x < pw; x++ {
			top := buf[(cy*2)*pw+x]
			var bot px
			if cy*2+1 < ph {
				bot = buf[(cy*2+1)*pw+x]
			}
			switch {
			case top.on && bot.on:
				s.cv.SetCellBG(ox+x, oy+cy, '▀', top.col, bot.col, 0.55)
			case top.on:
				s.cv.SetCell(ox+x, oy+cy, '▀', top.col, 0.55)
			case bot.on:
				s.cv.SetCell(ox+x, oy+cy, '▄', bot.col, 0.55)
			}
		}
	}
	s.drawTagline()
}

func (s *Scene) drawTitleCompact() {
	t := s.t
	runes := []rune("S P E C H T   L A B S")
	ox := (s.lay.w - len(runes)) / 2
	glyphs := []rune("!<>-_\\/[]{}=+*^?#░▒")
	for i, ch := range runes {
		if ch == ' ' {
			continue
		}
		resolve := tTitleStart + 0.2 + float64(i)*0.04
		col := r.Ramp(s.th.Title, 0.5)
		if t < resolve {
			ch = glyphs[int(hash(uint32(i), uint32(t*20), 5)*float64(len(glyphs)))%len(glyphs)]
			col = col.Scale(0.5)
		}
		s.cv.SetCell(ox+i, s.lay.titleY, ch, col, 0.6)
	}
	s.drawTagline()
}

func (s *Scene) drawTagline() {
	t := s.t
	if t < tTagStart {
		return
	}
	tag := []rune(s.opt.Tagline)
	n := min(len(tag), int((t-tTagStart)*38))
	ox := (s.lay.w - len(tag)) / 2
	s.cv.Text(ox, s.lay.tagY, string(tag[:n]), s.th.Text.Scale(0.85), 0.25)
	if math.Mod(t, 0.9) < 0.5 {
		s.cv.SetCell(ox+n, s.lay.tagY, '█', s.th.Text.Scale(0.85), 0.3)
	}
}

// ------------------------------------------------------------------- FX --

// drawCRT: a horizontal line snaps open, then the picture blooms vertically.
func (s *Scene) drawCRT() {
	t := s.t
	if t >= tCRTEnd {
		return
	}
	w, h := s.lay.w, s.lay.h
	white := r.RGB{R: 1, G: 1, B: 1}
	mid := h / 2
	if t < 0.22 {
		half := int(ease(t/0.22) * float64(w) / 2)
		for x := w/2 - half; x < w/2+half; x++ {
			s.cv.SetCell(x, mid, '━', white, 1.5)
			s.cv.AddBG(x, mid, s.th.Grad[1].Scale(0.35))
		}
		return
	}
	p := (t - 0.22) / (tCRTEnd - 0.22)
	half := ease(p) * float64(h) / 2
	for y := range h {
		d := math.Abs(float64(y - mid))
		if d > half {
			continue
		}
		v := (1 - p) * (1 - d/(half+1)) * 0.7
		for x := range w {
			s.cv.AddBG(x, y, white.Lerp(s.th.Grad[1], 0.5).Scale(v))
		}
	}
}

// drawScanline: a faint bright band rolling down the screen.
func (s *Scene) drawScanline() {
	h := s.lay.h
	y := int(math.Mod(s.t*0.18, 1.3) * float64(h))
	for d := -1; d <= 1; d++ {
		v := 0.035 * (1 - math.Abs(float64(d))*0.5)
		for x := 0; x < s.lay.w; x++ {
			s.cv.AddBG(x, y+d, s.th.Grad[2].Scale(v))
		}
	}
}

// glitch: brief horizontal tearing every few seconds once the scene holds.
func (s *Scene) glitch() {
	t := s.t
	if t < tHold+1 {
		return
	}
	const period = 4.3
	cyc := math.Floor((t - tHold) / period)
	local := math.Mod(t-tHold, period)
	if local > 0.16 {
		return
	}
	f := s.frame
	k := uint32(cyc)
	step := uint32(local * 30)
	for b := range 3 {
		y0 := int(hash(k, uint32(b), 21) * float64(f.H))
		bh := 1 + int(hash(k, uint32(b), 22)*3)
		shift := int((hash(k, uint32(b), 23+step) - 0.5) * 14)
		for y := y0; y < min(f.H, y0+bh); y++ {
			row := make([]r.Cell, f.W)
			copy(row, f.Cells[y*f.W:(y+1)*f.W])
			for x := 0; x < f.W; x++ {
				src := row[((x-shift)%f.W+f.W)%f.W]
				src.FG = src.FG.Mul(r.RGB{R: 0.6, G: 1.2, B: 1.3})
				src.BG = src.BG.Add(s.th.Grad[2].Scale(0.06))
				f.Cells[y*f.W+x] = src
			}
		}
	}
}
