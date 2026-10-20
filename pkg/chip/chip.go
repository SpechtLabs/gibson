// Package chip is a small tracker and sound chip in the spirit of the C64's
// SID and the NES's APU: pulse waves with pulse-width sweeps, a 4-bit
// stepped triangle, a sawtooth and LFSR noise, driven 50 times a second
// like a PAL demo's music routine, so arpeggios buzz at the speed they did
// in 1989. It mixes down to 8-bit unsigned mono PCM, and [Stream] sends that
// as a WAV in real time, for anything that plays one from a pipe.
//
// A [Song] is a set of channels, each a [Track] of rows written in tracker
// notation (see [Parse]) and played by [Instrument]s.
package chip

import (
	"errors"
	"fmt"
)

// TickRate is how many times a second the player steps: PAL's vertical
// blank, which demo music routines ran on.
const TickRate = 50

// Wave is an oscillator's waveform.
type Wave uint8

// The waveforms.
const (
	Pulse    Wave = iota // a square wave with a variable duty cycle
	Triangle             // the NES's 4-bit stepped triangle
	Saw                  // a sawtooth
	Noise                // a 15-bit LFSR clocked at the note's frequency
)

// Instrument is a voice's sound: a waveform, an ADSR envelope and the
// modulation on top. Times are in seconds.
type Instrument struct {
	Wave   Wave
	Volume float64

	Duty     float64 // pulse width, 0 to 1 (0.5 is square)
	PWMDepth float64 // how far the pulse width sweeps either way
	PWMRate  float64 // how fast, in Hz

	Attack, Decay, Release float64
	Sustain                float64 // level, 0 to 1
	Length                 float64 // release this long after the note starts; 0 holds it

	Vibrato      float64 // depth in semitones
	VibratoRate  float64 // Hz
	VibratoDelay float64 // how long the note sounds before vibrato sets in

	Sweep float64 // pitch slide in semitones per second
	Floor int     // the note a downward sweep stops at
	Glide float64 // time to slide from the note before, while it still sounds

	Note int // the note it plays as a kit hit (see [Kit])
}

// Channel is one voice of a song.
type Channel struct {
	Name  string
	Track Track
	Gain  float64
	Echo  float64 // how much is sent to the song's echo
}

// Song is channels played in lockstep, row by row.
type Song struct {
	Title    string
	Speed    int // ticks per row
	Loop     int // the row to go back to after the last
	Channels []Channel

	EchoRows     int     // echo delay, in rows
	EchoFeedback float64 // how much of the echo is echoed again
}

// Rows is the length of the song in rows.
func (s *Song) Rows() int {
	if len(s.Channels) == 0 {
		return 0
	}
	return len(s.Channels[0].Track)
}

// RowSeconds is how long a row lasts.
func (s *Song) RowSeconds() float64 { return float64(s.Speed) / TickRate }

// ErrEmptySong is returned for a song without rows.
var ErrEmptySong = errors.New("chip: song has no rows")

// Validate checks that the song can be played: its tracks are the same
// length, and its speed and loop point make sense.
func (s *Song) Validate() error {
	n := s.Rows()
	if n == 0 {
		return ErrEmptySong
	}
	for _, c := range s.Channels {
		if len(c.Track) != n {
			return fmt.Errorf("chip: channel %s has %d rows, %s has %d", c.Name, len(c.Track), s.Channels[0].Name, n)
		}
	}
	if s.Speed < 1 {
		return fmt.Errorf("chip: speed must be at least 1, got %d", s.Speed)
	}
	if s.Loop < 0 || s.Loop >= n {
		return fmt.Errorf("chip: loop row %d is outside the song's %d rows", s.Loop, n)
	}
	return nil
}
