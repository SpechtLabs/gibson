// Package scene assembles the gibson intro from the parts in pkg/: the boot
// log, the containment vessel with its field, the projector and the
// specimen it materializes, the HUD and the title, all in one theme, on one
// timeline.
package scene

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/components/boot"
	"github.com/spechtlabs/gibson/pkg/components/containment"
	"github.com/spechtlabs/gibson/pkg/components/crt"
	"github.com/spechtlabs/gibson/pkg/components/dust"
	"github.com/spechtlabs/gibson/pkg/components/glitch"
	"github.com/spechtlabs/gibson/pkg/components/hud"
	"github.com/spechtlabs/gibson/pkg/components/poweron"
	"github.com/spechtlabs/gibson/pkg/components/projector"
	"github.com/spechtlabs/gibson/pkg/components/specimen"
	"github.com/spechtlabs/gibson/pkg/components/specimen/bird"
	"github.com/spechtlabs/gibson/pkg/components/tagline"
	"github.com/spechtlabs/gibson/pkg/components/title"
	"github.com/spechtlabs/gibson/pkg/components/vessel"
	"github.com/spechtlabs/gibson/pkg/components/vessel/hex"
	r "github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/stage"
	"github.com/spechtlabs/gibson/pkg/theme"
	"github.com/spechtlabs/gibson/pkg/theme/synth"
)

// Timeline (seconds). Everything after tHold loops forever.
const (
	tCRTSnap     = 0.22
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

// DefaultTagline is the line typed under the title unless [WithTagline]
// says otherwise.
const DefaultTagline = "CONTAINMENT NOMINAL // WELCOME BACK, OPERATOR"

var bootLog = []string{
	"SPECHT LABS SYS/9000 :: BIOS 9.5.1",
	"MEM 65536K .................. OK",
	"NEURAL LATTICE .............. OK",
	"MOUNT /dev/specimen0 ........ OK",
	"LOAD containment.ko ......... OK",
	">> SPECIMEN SIGNAL DETECTED",
}

// Scene is the intro. Render draws the frame at a moment of it; a scene
// keeps no state between frames beyond its buffers.
type Scene struct {
	theme    theme.Theme
	specimen specimen.Specimen
	vessel   vessel.Vessel
	field    containment.Field
	tagline  string

	projector *projector.Projector
	dust      *dust.Dust
	overlays  []stage.Layer // drawn over the world once it's up, in order
	power     *poweron.PowerOn
	scan      crt.Scanline
	glitch    *glitch.Glitch

	cv    *r.Canvas
	frame *r.Frame
	st    stage.Stage
}

// Option configures a [Scene].
type Option func(*Scene)

// WithTheme sets the theme (default synth).
func WithTheme(t theme.Theme) Option { return func(s *Scene) { s.theme = t } }

// WithSpecimen sets what the projector materializes (default the bird).
func WithSpecimen(sp specimen.Specimen) Option { return func(s *Scene) { s.specimen = sp } }

// WithVessel sets the containment vessel (default the hex prism).
func WithVessel(v vessel.Vessel) Option { return func(s *Scene) { s.vessel = v } }

// WithContainment sets the containment field. By default every vessel
// comes with its own (see [Field]).
func WithContainment(f containment.Field) Option { return func(s *Scene) { s.field = f } }

// WithTagline sets the line typed under the title.
func WithTagline(text string) Option { return func(s *Scene) { s.tagline = text } }

// New assembles a scene.
func New(opts ...Option) *Scene {
	s := &Scene{tagline: DefaultTagline}
	for _, o := range opts {
		o(s)
	}
	if s.theme == nil {
		s.theme = synth.New()
	}
	if s.specimen == nil {
		s.specimen = bird.New()
	}
	if s.vessel == nil {
		s.vessel = hex.New()
	}
	if s.field == nil {
		s.field = Field(s.vessel)
	}

	s.projector = projector.New(projector.WithFloor(s.vessel.Floor()))
	s.dust = dust.New(dust.WithCount(70))
	s.overlays = []stage.Layer{
		hud.New(
			hud.WithStart(tHUD),
			hud.WithLabel("SPECHT LABS ▸ CONTAINMENT UNIT 07", tBootFade),
			hud.WithReadout(s.readout),
		),
		boot.New(boot.WithLines(bootLog...), boot.WithStart(tBootStart), boot.WithFadeAt(tBootFade)),
		title.New("SPECHT LABS", title.WithStart(tTitleStart), title.WithSheenFrom(tHold)),
		tagline.New(s.tagline, tagline.WithStart(tTagStart)),
	}
	s.power = poweron.New(poweron.WithSnap(tCRTSnap), poweron.WithBloom(tCRTEnd-tCRTSnap))
	s.glitch = glitch.New(glitch.WithFrom(tHold), glitch.WithDelay(1), glitch.WithPeriod(4.3))
	s.cv = r.NewCanvas(1, 1)
	return s
}

// Palette is the scene's theme's palette.
func (s *Scene) Palette() theme.Palette { return s.theme.Palette() }

// readout is what the HUD reports at t.
func (s *Scene) readout(t float64) hud.Readout {
	status := "ACQUIRING"
	switch {
	case t > tMatEnd+0.3:
		status = "CONTAINED"
	case t > tMatStart:
		status = "MATERIALIZING"
	}
	return hud.Readout{
		Designation: s.specimen.Designation(),
		Status:      status,
		Settled:     status == "CONTAINED",
		Warning:     t > tMatStart-0.2 && t < tMatEnd+0.6,
	}
}

// Render draws the scene at time t (seconds) into a w×h cell frame. The
// frame is reused by the next call.
func (s *Scene) Render(t float64, w, h int) *r.Frame {
	s.cv.Resize(w, h)
	st := &s.st
	*st = stage.Stage{
		Canvas:  s.cv,
		Palette: s.theme.Palette(),
		T:       t,
		// a slow turntable for the world, and a camera tilt that bobs
		Yaw:  0.5 + t*0.42,
		Tilt: 0.36 + 0.05*math.Sin(t*0.31),
	}
	st.Layout, st.Camera = layout(w, h)

	if t >= tBootStart {
		s.drawWorld(st)
		for _, l := range s.overlays {
			l.Draw(st)
		}
	}
	s.power.Draw(st)
	s.scan.Draw(st)

	s.frame = s.cv.Compose(crt.Vignette(st), 0.11, s.frame)
	s.glitch.Apply(st, s.frame)
	return s.frame
}

// drawWorld draws the 3D part. The specimen goes first and records its
// depth; everything after it is dimmed wherever it passes behind it.
func (s *Scene) drawWorld(st *stage.Stage) {
	t := st.T
	if t >= tMatStart {
		reveal := anim.Ease(anim.Phase(t, tMatStart, tMatEnd))
		s.projector.Materialize(st, s.specimen, reveal)
		s.cv.DilateDepth()
		s.cv.Occlude = behindDim
		s.projector.ScanRing(st, s.specimen, reveal)
		s.projector.Beam(st, anim.Ease(anim.Phase(t, tMatStart, tMatStart+0.6)), t > tHold)
	}
	s.cv.Occlude = behindDim
	defer func() { s.cv.Occlude = 0 }()
	s.projector.FloorRings(st, anim.Ease(anim.Phase(t, 1.1, 2.4)))

	build := anim.Ease(anim.Phase(t, tVesselStart, tVesselEnd))
	open := anim.Spring(anim.Phase(t, tSlideStart, tSlideStart+1.6))
	if t > tHold {
		open *= 1 - 0.08*(0.5-0.5*math.Cos((t-tHold)*0.9))
	}
	s.vessel.Draw(st, vessel.Progress{
		Build:  build,
		Unseal: anim.Phase(t, tCutStart, tCutEnd),
		Open:   open,
		Idle:   t > tHold,
	})
	s.projector.Pad(st, build)

	if in := anim.Ease(anim.Phase(t, tMatEnd-0.3, tMatEnd+0.5)); in > 0 {
		s.field.Draw(st, in)
	}
	s.dust.Draw(st, anim.Ease(anim.Phase(t, tMatStart+0.8, tMatEnd+1)))
}

// layout places the title and tagline at the bottom and fits the camera to
// the world above them.
func layout(w, h int) (stage.Layout, r.Camera) {
	l := stage.Layout{W: w, H: h, Compact: w < 64 || h < 22, TitleScale: 1}
	if w >= 150 && h >= 46 {
		l.TitleScale = 2
	}
	titleRows := (7*l.TitleScale + 1) / 2
	if l.Compact {
		titleRows = 1
	}
	l.TagY = h - 5
	l.TitleY = l.TagY - 1 - titleRows
	if l.Compact {
		l.TagY = h - 2
		l.TitleY = h - 4
	}
	top, bot := 8.0, float64(l.TitleY*4-2)
	areaH := bot - top
	// world extents we want to fit: y in [-1.05, 1.75], x in [-2.1, 2.1]
	f := math.Min(areaH/3.7, float64(w*2)/5.4)
	return l, r.Camera{Dist: 6, F: f * 6, CX: float64(w), CY: (top+bot)/2 + 0.45*f}
}
