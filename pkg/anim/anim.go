// Package anim holds the small timing helpers every animated part shares:
// clamping, phases of a timeline, easing curves and a stateless hash for
// deterministic noise.
package anim

import "math"

// Clamp limits v to [lo, hi].
func Clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// Phase is how far t is through [a, b], clamped to [0, 1].
func Phase(t, a, b float64) float64 { return Clamp((t-a)/(b-a), 0, 1) }

// Ease is a cubic ease-out.
func Ease(x float64) float64 { return 1 - (1-x)*(1-x)*(1-x) }

// Spring is a damped overshoot from 0 to 1.
func Spring(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	return 1 - math.Exp(-5*x)*math.Cos(7.5*x)*(1-x)
}

// Hash maps three integers to a well-mixed value in [0, 1], so noise can be
// a pure function of time and index rather than state.
func Hash(a, b, c uint32) float64 {
	x := a*0x9E3779B1 ^ b*0x85EBCA77 ^ c*0xC2B2AE3D ^ 0x27d4eb2f
	x ^= x >> 15
	x *= 0x2C1B3C6D
	x ^= x >> 12
	x *= 0x297A2D39
	x ^= x >> 15
	return float64(x) / float64(^uint32(0))
}
