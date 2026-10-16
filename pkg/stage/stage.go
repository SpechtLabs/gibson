// Package stage is what every part of the scene draws onto: the canvas, the
// camera and the moment in time, plus a pen that draws world-space lines
// with perspective and depth cueing.
//
// A scene fills in a Stage once per frame and hands it to each part in turn.
// The parts themselves are stateless about time: they read it from the
// Stage, so the same frame always renders the same way.
package stage

import (
	"math"

	"github.com/spechtlabs/gibson/pkg/anim"
	"github.com/spechtlabs/gibson/pkg/render"
	"github.com/spechtlabs/gibson/pkg/theme"
)

// Stage is the per-frame drawing context.
type Stage struct {
	Canvas  *render.Canvas
	Palette theme.Palette
	Camera  render.Camera
	Layout  Layout

	T    float64 // seconds into the sequence
	Yaw  float64 // turntable angle of the world, around Y
	Tilt float64 // camera tilt, around X
}

// Layout is where the screen-space parts go.
type Layout struct {
	W, H       int  // cells
	Compact    bool // too small for the full HUD and the block title
	TitleY     int
	TitleScale int
	TagY       int
}

// Layer is a part drawn straight onto the stage, such as an overlay.
type Layer interface {
	Draw(st *Stage)
}

// View turns a world-space point into camera space.
func (s *Stage) View(v render.Vec3) render.Vec3 { return v.RotX(s.Tilt) }

// Project maps a world-space point to canvas sub-pixels.
func (s *Stage) Project(v render.Vec3) (x, y float64, ok bool) {
	return s.Camera.Project(s.View(v))
}

// Turntable rotates a model-space point by the world's yaw.
func (s *Stage) Turntable(v render.Vec3) render.Vec3 { return v.RotY(s.Yaw) }

// HeightColor colors a camera-space point by its height along the theme's
// gradient.
func (s *Stage) HeightColor(v render.Vec3) render.RGB {
	return render.Ramp(s.Palette.Grad, anim.Clamp((1.3-v.Y)/2.6, 0, 1))
}

// Identity is the model transform for parts already in world space.
func Identity(v render.Vec3) render.Vec3 { return v }

// Solid colors every point c.
func Solid(c render.RGB) func(render.Vec3) render.RGB {
	return func(render.Vec3) render.RGB { return c }
}

// Pen draws model-space segments: Xf takes them to world space, Color picks
// the color at a camera-space point, and I is the intensity.
type Pen struct {
	St    *Stage
	Xf    func(render.Vec3) render.Vec3
	Color func(render.Vec3) render.RGB
	I     float64
}

// Pen returns a pen on s.
func (s *Stage) Pen(xf func(render.Vec3) render.Vec3, color func(render.Vec3) render.RGB, i float64) Pen {
	return Pen{St: s, Xf: xf, Color: color, I: i}
}

// Seg draws the first frac of the segment a-b. Nearer points burn brighter.
func (p Pen) Seg(a, b render.Vec3, frac float64) {
	if frac <= 0 {
		return
	}
	if frac < 1 {
		b = a.Lerp(b, frac)
	}
	wa, wb := p.St.View(p.Xf(a)), p.St.View(p.Xf(b))
	cam := p.St.Camera
	x0, y0, ok0 := cam.Project(wa)
	x1, y1, ok1 := cam.Project(wb)
	if !ok0 || !ok1 {
		return
	}
	depth := func(v render.Vec3) float64 { return anim.Clamp(1-v.Z*0.22, 0.4, 1.3) }
	p.St.Canvas.Line(x0, y0, wa.Z, x1, y1, wb.Z, p.Color(wa), p.Color(wb), p.I*depth(wa), p.I*depth(wb))
}

// Poly draws the first frac of the polyline pts.
func (p Pen) Poly(pts []render.Vec3, frac float64) {
	n := len(pts) - 1
	for i := range n {
		f := anim.Clamp(frac*float64(n)-float64(i), 0, 1)
		p.Seg(pts[i], pts[i+1], f)
	}
}

// Circle draws a horizontal circle of radius rad at height y.
func (p Pen) Circle(y, rad float64, seg int) {
	pts := make([]render.Vec3, 0, seg+1)
	for k := 0; k <= seg; k++ {
		a := float64(k) / float64(seg) * 2 * math.Pi
		pts = append(pts, render.V(math.Cos(a)*rad, y, math.Sin(a)*rad))
	}
	p.Poly(pts, 1)
}
