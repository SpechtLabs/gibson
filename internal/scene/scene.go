package scene

import (
	"math"

	r "github.com/spechtlabs/gibson/internal/render"
	"github.com/spechtlabs/gibson/internal/theme"
)

// Options picks what a scene shows.
type Options struct {
	Theme    theme.Theme
	Specimen string // "bird" or "blob"
	Vessel   string // "hex" or "dodeca"
	Tagline  string
}

// The Options.Specimen and Options.Vessel names a scene knows.
var (
	Specimens = []string{"bird", specimenBlob}
	Vessels   = []string{"hex", "dodeca"}
)

const specimenBlob = "blob"

// Timeline (seconds). Everything after tHold loops forever.
const (
	tCRTEnd      = 0.55
	tBootStart   = 0.45
	tBootFade    = 2.9
	tHUD         = 1.0
	tVesselStart = 1.5
	tVesselEnd   = 2.7
	tCutStart    = 2.7
	tCutEnd      = 3.2
	tSlideStart  = 3.15
	tMatStart    = 3.6
	tMatEnd      = 5.1
	tTitleStart  = 4.7
	tTagStart    = 6.0
	tHold        = 7.0
)

// behindDim scales the brightness of anything passing behind the specimen.
const behindDim = 0.3

// Scene renders the intro sequence; Render draws the frame at a time.
type Scene struct {
	opt   Options
	th    theme.Theme
	cv    *r.Canvas
	frame *r.Frame
	ves   *vessel
	bird  []strand
	parts []particle

	// per-frame state
	t   float64
	lay layout
}

type particle struct {
	pos   r.Vec3
	dir   r.Vec3
	line  bool
	speed float64
	phase float64
}

// New builds a scene for opt.
func New(opt Options) *Scene {
	s := &Scene{opt: opt, th: opt.Theme, cv: r.NewCanvas(1, 1), ves: newVessel(opt.Vessel), bird: birdModel()}
	for i := range 70 {
		a := hash(uint32(i), 1, 0) * 2 * math.Pi
		rad := 1.5 + hash(uint32(i), 2, 0)*2.4
		s.parts = append(s.parts, particle{
			pos:   r.V(math.Cos(a)*rad, -1.3+hash(uint32(i), 3, 0)*3.4, math.Sin(a)*rad),
			dir:   r.V(hash(uint32(i), 4, 0)-0.5, hash(uint32(i), 5, 0)-0.5, hash(uint32(i), 6, 0)-0.5).Norm(),
			line:  hash(uint32(i), 7, 0) < 0.55,
			speed: 0.05 + hash(uint32(i), 8, 0)*0.12,
			phase: hash(uint32(i), 9, 0) * 10,
		})
	}
	return s
}

type layout struct {
	w, h       int
	compact    bool
	cam        r.Camera
	titleY     int
	titleScale int
	tagY       int
}

func (s *Scene) doLayout(w, h int) {
	l := layout{w: w, h: h}
	l.compact = w < 64 || h < 22
	l.titleScale = 1
	if w >= 150 && h >= 46 {
		l.titleScale = 2
	}
	titleRows := (7*l.titleScale + 1) / 2
	if l.compact {
		titleRows = 1
	}
	l.tagY = h - 5
	l.titleY = l.tagY - 1 - titleRows
	if l.compact {
		l.tagY = h - 2
		l.titleY = h - 4
	}
	top, bot := 8.0, float64(l.titleY*4-2)
	areaH := bot - top
	// world extents we want to fit: y in [-1.05, 1.75], x in [-2.1, 2.1]
	f := math.Min(areaH/3.7, float64(w*2)/5.4)
	l.cam = r.Camera{Dist: 6, F: f * 6, CX: float64(w), CY: (top+bot)/2 + 0.45*f}
	s.lay = l
}

// Render draws the scene at time t (seconds) into a w×h cell frame.
func (s *Scene) Render(t float64, w, h int) *r.Frame {
	s.t = t
	s.cv.Resize(w, h)
	s.doLayout(w, h)

	if t >= tBootStart {
		s.drawWorld()
		s.drawHUD()
		s.drawBoot()
		s.drawTitle()
	}
	s.drawCRT()
	s.drawScanline()

	bg := s.th.BG
	s.frame = s.cv.Compose(func(x, y int) r.RGB {
		dx := (float64(x)/float64(w) - 0.5) * 2
		dy := (float64(y)/float64(h) - 0.5) * 2
		v := 1 - 0.7*math.Min(1, (dx*dx+dy*dy)*0.55)
		if y%2 == 1 {
			v *= 0.82
		}
		return bg.Scale(v)
	}, 0.11, s.frame)
	s.glitch()
	return s.frame
}

// ----------------------------------------------------------------- world --

// pen draws world-space segments with depth cueing and height gradient.
type pen struct {
	s     *Scene
	xf    func(r.Vec3) r.Vec3 // model -> world
	color func(world r.Vec3) r.RGB
	i     float64
}

func (s *Scene) view(v r.Vec3) r.Vec3 {
	// camera orbit: slow continuous yaw is applied by models; here only tilt + bob
	tilt := 0.36 + 0.05*math.Sin(s.t*0.31)
	return v.RotX(tilt)
}

func (p pen) seg(a, b r.Vec3, frac float64) {
	if frac <= 0 {
		return
	}
	if frac < 1 {
		b = a.Lerp(b, frac)
	}
	wa, wb := p.s.view(p.xf(a)), p.s.view(p.xf(b))
	cam := p.s.lay.cam
	x0, y0, ok0 := cam.Project(wa)
	x1, y1, ok1 := cam.Project(wb)
	if !ok0 || !ok1 {
		return
	}
	depth := func(v r.Vec3) float64 { return clamp(1-v.Z*0.22, 0.4, 1.3) }
	p.s.cv.Line(x0, y0, wa.Z, x1, y1, wb.Z, p.color(wa), p.color(wb), p.i*depth(wa), p.i*depth(wb))
}

func (p pen) poly(pts []r.Vec3, frac float64) {
	n := len(pts) - 1
	for i := range n {
		f := clamp(frac*float64(n)-float64(i), 0, 1)
		p.seg(pts[i], pts[i+1], f)
	}
}

func (s *Scene) heightColor(v r.Vec3) r.RGB {
	return r.Ramp(s.th.Grad, clamp((1.3-v.Y)/2.6, 0, 1))
}

func (s *Scene) yaw() float64 { return 0.5 + s.t*0.42 }

func (s *Scene) drawWorld() {
	t := s.t
	yaw := s.yaw()
	rot := func(v r.Vec3) r.Vec3 { return v.RotY(yaw) }

	// the specimen goes first and records its depth; everything after it is
	// dimmed wherever it passes behind the specimen
	s.drawSpecimen()
	s.cv.Occlude = behindDim
	defer func() { s.cv.Occlude = 0 }()
	s.drawOrbits()

	// ---- vessel
	drawIn := ease(phase(t, tVesselStart, tVesselEnd))
	cutIn := phase(t, tCutStart, tCutEnd)
	slide := spring(phase(t, tSlideStart, tSlideStart+1.6))
	if t > tHold {
		slide *= 1 - 0.08*(0.5-0.5*math.Cos((t-tHold)*0.9))
	}
	if s.ves.panels != nil {
		s.drawPanels(rot, drawIn, cutIn, slide)
	} else {
		s.drawSliced(rot, drawIn, cutIn, slide)
	}
	// base pad (the hatched plate under the specimen)
	vp := pen{s: s, xf: rot, color: s.heightColor, i: 0.45}
	for k := range 6 {
		z := -0.32 + float64(k)*0.128
		vp.seg(r.V(-0.45, -s.ves.halfH+0.02, z), r.V(0.45, -s.ves.halfH+0.02, z), drawIn)
	}

	s.drawRings()
	s.drawParticles()
}

// drawPanels opens the vessel as an exploded view: faces push out along
// their normals (the top ones lift more, the floor stays), shrink and twist
// a little so clean gaps open up.
func (s *Scene) drawPanels(rot func(r.Vec3) r.Vec3, drawIn, cutIn, open float64) {
	t := s.t
	up := r.V(0, 1, 0)
	for i, p := range s.ves.panels {
		lift := (p.normal.Y + 1) / 2 // 0 for the floor, 1 for the lid
		fi := float64(i)
		off := p.normal.Scale(0.8 * open * lift).Add(up.Scale(0.45 * open * lift * lift))
		if t > tHold {
			off = off.Add(p.normal.Scale(0.04 * lift * math.Sin(t*1.3+fi)))
		}
		shrink := 1 - 0.3*open*lift
		twist := 0.5 * open * lift
		if i%2 == 1 {
			twist = -twist
		}
		xf := func(v r.Vec3) r.Vec3 {
			d := rotAxis(v.Sub(p.center), p.normal, twist).Scale(shrink)
			return rot(p.center.Add(off).Add(d))
		}
		// seams flare while the panels unlatch; shared edges overlap while
		// closed, so each face draws at roughly half strength then
		flare := 1 + 0.8*math.Sin(cutIn*math.Pi)
		pp := pen{s: s, xf: xf, color: s.heightColor, i: 0.4 * flare * (1 + 0.6*clamp(open*4, 0, 1))}
		pp.poly(p.pts, drawIn)
	}
}

// drawSliced cuts the vessel along a tilted plane and slides the lid off.
func (s *Scene) drawSliced(rot func(r.Vec3) r.Vec3, drawIn, cutIn, slide float64) {
	// the prism's front faces sit right over the specimen, so even lines in
	// front of it are dimmed somewhat (and the dense hatching fully)
	s.cv.Occlude, s.cv.OccludeFront = 0.18, 0.35
	defer func() { s.cv.Occlude, s.cv.OccludeFront = behindDim, 0 }()
	off := s.ves.slide.Scale(0.8 * slide).Add(s.ves.n.Scale(0.2 * slide))
	vp := pen{s: s, xf: rot, color: s.heightColor, i: 0.7}
	up := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return rot(v.Add(off)) }, color: s.heightColor, i: 0.7}

	lower, upper := s.ves.split()
	for _, e := range lower {
		vp.seg(e[0], e[1], drawIn)
	}
	for _, e := range upper {
		up.seg(e[0], e[1], drawIn)
	}
	if cutIn > 0 {
		cut := append(append([]r.Vec3{}, s.ves.cut...), s.ves.cut[0])
		vp.i = 1.15
		vp.poly(cut, cutIn)
		up.i = 1.15
		if slide > 0.02 {
			up.poly(cut, 1)
		}
		hs := s.ves.hatches(9)
		vp.i, up.i = 0.5, 0.5
		s.cv.OccludeFront = 0.18
		for k, hl := range hs {
			f := clamp(cutIn*1.6-float64(k)*0.07, 0, 1)
			vp.seg(hl[0], hl[1], f)
			if slide > 0.05 {
				up.seg(hl[0], hl[1], f)
			}
		}
	}
}

func (s *Scene) drawOrbits() {
	t := s.t
	in := ease(phase(t, 1.1, 2.4))
	if in <= 0 {
		return
	}
	p := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v }, color: func(r.Vec3) r.RGB { return s.th.Orbit }, i: 0.6}
	type ring struct{ y, rad, speed, span, ph float64 }
	rings := []ring{
		{-1.12, 1.5, 0.35, 1.7, 0}, {-1.12, 1.8, -0.22, 1.25, 1.3}, {-1.14, 2.15, 0.15, 1.55, 2.2},
	}
	for _, rg := range rings {
		const seg = 64
		start := rg.ph + rg.speed*t
		span := rg.span * math.Pi * in
		rad := rg.rad * (0.6 + 0.4*in)
		var pts []r.Vec3
		for k := 0; k <= seg; k++ {
			a := start + span*float64(k)/seg
			pts = append(pts, r.V(math.Cos(a)*rad, rg.y, math.Sin(a)*rad))
		}
		p.poly(pts, 1)
	}
}

func (s *Scene) specimenXf() func(r.Vec3) r.Vec3 {
	t := s.t
	sway := 0.45*math.Sin(t*0.45) + 0.1*math.Sin(t*1.3)
	bob := 0.05 * math.Sin(t*1.7)
	scale := 1.12
	if s.opt.Specimen == specimenBlob {
		scale = 0.9
	}
	return func(v r.Vec3) r.Vec3 {
		return v.Scale(scale).RotY(sway + math.Pi*0.08).Add(r.V(0, -0.02+bob, 0))
	}
}

func (s *Scene) drawSpecimen() {
	t := s.t
	if t < tMatStart {
		return
	}
	reveal := ease(phase(t, tMatStart, tMatEnd))
	level := -1.0 + reveal*2.05 // model-space scan height
	xf := s.specimenXf()
	pulse := 1.1 + 0.12*math.Sin(t*3.1)
	hot := s.th.Specimen
	var model []strand
	if s.opt.Specimen == specimenBlob {
		model = blobModel(t)
	} else {
		model = s.bird
	}
	p := pen{s: s, xf: xf, i: pulse}
	specCol := func(r.Vec3) r.RGB { return hot }
	q := uint32(t * 20) // jitter changes 20×/s
	s.cv.MarkMask = true
	for li, st := range model {
		pl := st.pts
		p.i = pulse * st.i
		p.color = specCol
		if st.accent {
			p.color = func(r.Vec3) r.RGB { return s.th.Eye }
		}
		for i := 0; i+1 < len(pl); i++ {
			a, b := pl[i], pl[i+1]
			if a.Y > level && b.Y > level {
				continue
			}
			// clip at the scan level
			if b.Y > level {
				b = a.Lerp(b, (level-a.Y)/(b.Y-a.Y))
			} else if a.Y > level {
				a = b.Lerp(a, (level-b.Y)/(a.Y-b.Y))
			}
			// noisy near the scan front
			if reveal < 1 && level-a.Y < 0.15 {
				j := (hash(uint32(li), uint32(i), q) - 0.5) * 0.12
				a.X += j
				b.X += j
			}
			p.seg(a, b, 1)
		}
	}
	s.cv.MarkMask = false
	s.cv.DilateDepth()
	s.cv.Occlude = behindDim
	defer func() { s.cv.Occlude = 0 }()
	// scan ring at the reveal front
	if reveal < 1 {
		rp := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v.RotY(s.yaw()) }, color: func(r.Vec3) r.RGB { return s.th.Beam }, i: 1.3}
		y := xf(r.V(0, level, 0)).Y
		s.circle(rp, y, 1.0, 48)
	}
	// projector beam
	bi := ease(phase(t, tMatStart, tMatStart+0.6)) * (0.55 + 0.25*hash(q, 77, 1))
	if t > tHold {
		bi *= 0.6
	}
	if bi > 0 {
		bp := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v.RotY(-s.yaw() * 0.5) }, color: func(r.Vec3) r.RGB { return s.th.Beam }, i: bi}
		src := r.V(0, -s.ves.halfH+0.02, 0)
		for k := range 6 {
			a := float64(k) / 6 * 2 * math.Pi
			bp.seg(src, r.V(math.Cos(a)*0.3, -0.3, math.Sin(a)*0.3), 1)
		}
	}
}

func (s *Scene) circle(p pen, y, rad float64, seg int) {
	var pts []r.Vec3
	for k := 0; k <= seg; k++ {
		a := float64(k) / float64(seg) * 2 * math.Pi
		pts = append(pts, r.V(math.Cos(a)*rad, y, math.Sin(a)*rad))
	}
	p.poly(pts, 1)
}

// containment rings around the specimen
func (s *Scene) drawRings() {
	t := s.t
	in := ease(phase(t, tMatEnd-0.3, tMatEnd+0.5))
	if in <= 0 {
		return
	}
	if s.ves.panels != nil {
		s.drawOrbiters(in)
		return
	}
	// like the prism, the rings fade over the specimen in front and behind
	s.cv.Occlude, s.cv.OccludeFront = 0.18, 0.35
	defer func() { s.cv.Occlude, s.cv.OccludeFront = behindDim, 0 }()
	p := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v }, color: func(r.Vec3) r.RGB { return s.th.Rings }, i: 0.9 * in}
	for i := range 3 {
		fi := float64(i)
		base := -0.6 + fi*0.55
		y := base + 0.07*math.Sin(t*1.2+fi*1.7)
		// rings drop in from above
		y += (1 - in) * (1.2 + fi*0.3)
		s.circle(p, y, 0.98+0.03*math.Sin(t*2+fi), 56)
	}
}

// drawOrbiters: the dodecahedron's containment field. Three nodes circle
// the specimen on tilted paths like an atom model, each trailing a fading
// comet tail instead of a full ring.
func (s *Scene) drawOrbiters(in float64) {
	t := s.t
	type orbit struct{ tiltX, tiltZ, speed, phase float64 }
	orbits := []orbit{{1.15, 0.25, 1.6, 0}, {-0.95, 0.9, -1.25, 2.1}, {0.2, -1.2, 1.05, 4.2}}
	const tail, steps = 2.2, 36 // tail length (rad), segments
	for _, o := range orbits {
		rad := 1.35 * (0.25 + 0.75*in)
		pt := func(a float64) r.Vec3 {
			return r.V(math.Cos(a)*rad, 0, math.Sin(a)*rad).RotX(o.tiltX).RotZ(o.tiltZ).Add(r.V(0, -0.02, 0))
		}
		head := o.phase + o.speed*t
		dir := 1.0
		if o.speed < 0 {
			dir = -1
		}
		p := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v }, color: func(r.Vec3) r.RGB { return s.th.Rings }}
		for k := range steps {
			f0, f1 := float64(k)/steps, float64(k+1)/steps
			p.i = 1.4 * in * math.Pow(1-f0, 1.1)
			p.seg(pt(head-dir*tail*f0), pt(head-dir*tail*f1), 1)
		}
		// the node itself: a short bright tick across the path
		h := pt(head)
		p.i = 1.8 * in
		p.color = func(r.Vec3) r.RGB { return s.th.Rings.Lerp(r.RGB{R: 1, G: 1, B: 1}, 0.5) }
		p.seg(h.Add(r.V(0, 0.06, 0)), h.Sub(r.V(0, 0.06, 0)), 1)
		p.seg(h.Add(r.V(0.06, 0, 0)), h.Sub(r.V(0.06, 0, 0)), 1)
	}
}

func (s *Scene) drawParticles() {
	t := s.t
	in := ease(phase(t, tMatStart+0.8, tMatEnd+1))
	if in <= 0 {
		return
	}
	yaw := s.yaw() * 0.35
	for i, pt := range s.parts {
		pos := pt.pos
		pos.Y = math.Mod(pos.Y+1.3+t*pt.speed, 3.4) - 1.3
		pos = pos.RotY(yaw)
		tw := 0.5 + 0.5*math.Sin(t*2.3+pt.phase)
		if pt.line {
			d := pt.dir.RotY(t * 0.4 * pt.speed * 8)
			p := pen{s: s, xf: func(v r.Vec3) r.Vec3 { return v }, color: func(r.Vec3) r.RGB { return s.th.Beam.Lerp(s.th.Orbit, 0.35) }, i: 0.4 * in * (0.5 + tw)}
			p.seg(pos.Sub(d.Scale(0.12)), pos.Add(d.Scale(0.12)), 1)
		} else {
			w := s.view(pos)
			x, y, ok := s.lay.cam.Project(w)
			if ok {
				s.cv.Plot(int(x), int(y), w.Z, s.th.Rings, 1.2*in*tw)
				s.cv.Plot(int(x)+1, int(y), w.Z, s.th.Rings, 0.6*in*tw)
			}
		}
		_ = i
	}
}

// ---------------------------------------------------------------- helpers --

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func phase(t, a, b float64) float64 { return clamp((t-a)/(b-a), 0, 1) }

func ease(x float64) float64 { return 1 - (1-x)*(1-x)*(1-x) }

// spring: damped overshoot 0 -> 1
func spring(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	return 1 - math.Exp(-5*x)*math.Cos(7.5*x)*(1-x)
}

func hash(a, b, c uint32) float64 {
	x := a*0x9E3779B1 ^ b*0x85EBCA77 ^ c*0xC2B2AE3D ^ 0x27d4eb2f
	x ^= x >> 15
	x *= 0x2C1B3C6D
	x ^= x >> 12
	x *= 0x297A2D39
	x ^= x >> 15
	return float64(x) / float64(^uint32(0))
}
