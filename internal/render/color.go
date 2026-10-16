package render

import "math"

// RGB is a linear-ish color with channels nominally in [0,1]. Values above 1
// are allowed while compositing and are clamped on output.
type RGB struct{ R, G, B float64 }

// Hex converts 0xRRGGBB to an RGB.
func Hex(h uint32) RGB {
	return RGB{float64(h>>16&0xff) / 255, float64(h>>8&0xff) / 255, float64(h&0xff) / 255}
}

func (c RGB) Add(o RGB) RGB       { return RGB{c.R + o.R, c.G + o.G, c.B + o.B} }
func (c RGB) Scale(f float64) RGB { return RGB{c.R * f, c.G * f, c.B * f} }
func (c RGB) Mul(o RGB) RGB       { return RGB{c.R * o.R, c.G * o.G, c.B * o.B} }
func (c RGB) Lerp(o RGB, t float64) RGB {
	return RGB{c.R + (o.R-c.R)*t, c.G + (o.G-c.G)*t, c.B + (o.B-c.B)*t}
}
func (c RGB) Luma() float64 { return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B }

// Bytes clamps and converts to 8-bit channels.
func (c RGB) Bytes() (uint8, uint8, uint8) {
	return to8(c.R), to8(c.G), to8(c.B)
}

func to8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(math.Round(v * 255))
}

// Ramp samples a multi-stop gradient at t in [0,1].
func Ramp(stops []RGB, t float64) RGB {
	if len(stops) == 0 {
		return RGB{}
	}
	if t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	f := t * float64(len(stops)-1)
	i := int(f)
	return stops[i].Lerp(stops[i+1], f-float64(i))
}
