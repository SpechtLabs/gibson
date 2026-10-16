package render

import "math"

// Vec3 is a point or direction in scene space.
type Vec3 struct{ X, Y, Z float64 }

// V is shorthand for a Vec3 literal.
func V(x, y, z float64) Vec3 { return Vec3{x, y, z} }

func (a Vec3) Add(b Vec3) Vec3             { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3             { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(f float64) Vec3        { return Vec3{a.X * f, a.Y * f, a.Z * f} }
func (a Vec3) Dot(b Vec3) float64          { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Len() float64                { return math.Sqrt(a.Dot(a)) }
func (a Vec3) Lerp(b Vec3, t float64) Vec3 { return a.Add(b.Sub(a).Scale(t)) }
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}
func (a Vec3) Norm() Vec3 {
	l := a.Len()
	if l == 0 {
		return a
	}
	return a.Scale(1 / l)
}

func (a Vec3) RotY(t float64) Vec3 {
	s, c := math.Sincos(t)
	return Vec3{a.X*c + a.Z*s, a.Y, -a.X*s + a.Z*c}
}

func (a Vec3) RotX(t float64) Vec3 {
	s, c := math.Sincos(t)
	return Vec3{a.X, a.Y*c - a.Z*s, a.Y*s + a.Z*c}
}

func (a Vec3) RotZ(t float64) Vec3 {
	s, c := math.Sincos(t)
	return Vec3{a.X*c - a.Y*s, a.X*s + a.Y*c, a.Z}
}

// Camera is a simple perspective camera looking down +Z from (0,0,-Dist).
// Output coordinates are in canvas sub-pixels.
type Camera struct {
	CX, CY float64 // screen-space center
	F      float64 // focal length in sub-pixels
	Dist   float64
}

func (c Camera) Project(v Vec3) (x, y float64, ok bool) {
	z := v.Z + c.Dist
	if z < 0.05 {
		return 0, 0, false
	}
	return c.CX + v.X*c.F/z, c.CY - v.Y*c.F/z, true
}
