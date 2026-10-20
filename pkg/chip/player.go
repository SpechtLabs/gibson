package chip

import "math"

// Player plays a song as 8-bit unsigned mono PCM, looping forever.
type Player struct {
	song   *Song
	rate   float64
	voices []voice

	row, tick  int
	tickLeft   float64 // samples until the next tick
	tickLength float64

	echo    []float64
	echoPos int
	lp      float64 // output filter state
	lpCoef  float64
}

// NewPlayer returns a player for s at rate samples per second.
func NewPlayer(s *Song, rate int) (*Player, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	p := &Player{
		song:       s,
		rate:       float64(rate),
		voices:     make([]voice, len(s.Channels)),
		tickLength: float64(rate) / TickRate,
		// the DAC's output filter: takes the edge off aliasing and the
		// noise channel without dulling the pulses
		lpCoef: 1 - math.Exp(-2*math.Pi*7500/float64(rate)),
	}
	for i := range p.voices {
		p.voices[i].lfsr = 1
	}
	if s.EchoRows > 0 {
		p.echo = make([]float64, int(float64(s.EchoRows*s.Speed)*p.tickLength))
	}
	return p, nil
}

// Read fills b with samples. It never runs out, so it never returns an
// error.
func (p *Player) Read(b []byte) (int, error) {
	for i := range b {
		if p.tickLeft <= 0 {
			p.step()
			p.tickLeft += p.tickLength
		}
		p.tickLeft--
		b[i] = quantize(p.sample())
	}
	return len(b), nil
}

// step runs one tick: a new row on the first tick of each, and the
// arpeggios on every one.
func (p *Player) step() {
	if p.tick == 0 {
		for i, c := range p.song.Channels {
			r := c.Track[p.row]
			switch {
			case r.Note == Off:
				p.voices[i].release()
			case r.Note > 0:
				p.voices[i].trigger(r)
			}
		}
	}
	for i := range p.voices {
		p.voices[i].arpTick++
	}
	p.tick++
	if p.tick == p.song.Speed {
		p.tick = 0
		p.row++
		if p.row == p.song.Rows() {
			p.row = p.song.Loop
		}
	}
}

// sample mixes the next sample, in -1 to 1.
func (p *Player) sample() float64 {
	dt := 1 / p.rate
	var mix, send float64
	for i, c := range p.song.Channels {
		v := p.voices[i].next(dt, p.rate) * c.Gain
		mix += v
		send += v * c.Echo
	}
	if p.echo != nil {
		d := p.echo[p.echoPos]
		mix += d
		p.echo[p.echoPos] = send + d*p.song.EchoFeedback
		p.echoPos = (p.echoPos + 1) % len(p.echo)
	}
	p.lp += p.lpCoef * (mix - p.lp)
	return math.Tanh(p.lp)
}

func quantize(x float64) byte {
	return byte(math.Round(127.5 + 127.5*max(-1, min(1, x))))
}

// Envelope stages.
const (
	idle = iota
	attack
	decay
	sustain
	release
)

type voice struct {
	inst    *Instrument
	vol     float64
	arp     []int
	arpTick int

	pitch, from float64 // the note, and the one glided from
	age         float64 // seconds since the note started
	phase       float64

	stage int
	env   float64

	lfsr  uint16
	noise float64
}

func (v *voice) trigger(r Row) {
	sounding := v.stage != idle && v.stage != release
	if v.inst == nil || r.Inst.Glide == 0 || !sounding {
		v.from = float64(r.Note)
	} else {
		v.from = v.current()
	}
	v.inst, v.vol, v.arp, v.arpTick = r.Inst, r.Vol, r.Arp, 0
	v.pitch, v.age = float64(r.Note), 0
	// the envelope restarts from where it is, so a retrigger doesn't click
	v.stage = attack
	if r.Inst.Wave == Noise || r.Inst.Sweep != 0 {
		v.phase = 0 // drums hit the same way every time
	}
}

func (v *voice) release() {
	if v.stage != idle {
		v.stage = release
	}
}

// current is the pitch sounding now, glide and sweep included.
func (v *voice) current() float64 {
	in := v.inst
	n := v.pitch
	if in.Glide > 0 && v.age < in.Glide {
		n = v.from + (v.pitch-v.from)*v.age/in.Glide
	}
	if in.Sweep != 0 {
		n += in.Sweep * v.age
		if in.Sweep < 0 {
			n = max(n, float64(in.Floor))
		}
	}
	return n
}

// next is the voice's next sample, dt seconds on.
func (v *voice) next(dt, rate float64) float64 {
	if v.stage == idle {
		return 0
	}
	in := v.inst
	v.age += dt
	v.envelope(dt)

	n := v.current()
	if len(v.arp) > 0 {
		n += float64(v.arp[v.arpTick%len(v.arp)])
	}
	if in.Vibrato > 0 && v.age > in.VibratoDelay {
		ramp := min(1, (v.age-in.VibratoDelay)/0.3)
		n += ramp * in.Vibrato * math.Sin(2*math.Pi*in.VibratoRate*(v.age-in.VibratoDelay))
	}
	inc := min(0.5, 440*math.Exp2((n-69)/12)/rate)

	var s float64
	switch in.Wave {
	case Pulse:
		duty := in.Duty + in.PWMDepth*math.Sin(2*math.Pi*in.PWMRate*v.age)
		s = pulse(v.phase, inc, max(0.05, min(0.95, duty)))
	case Triangle:
		s = triangle(v.phase)
	case Saw:
		s = 2*v.phase - 1 - blep(v.phase, inc)
	case Noise:
		s = v.clockNoise(inc * 2)
	}
	v.phase += inc
	v.phase -= math.Floor(v.phase)
	return s * v.env * v.vol * in.Volume
}

func (v *voice) envelope(dt float64) {
	in := v.inst
	if in.Length > 0 && v.age >= in.Length && v.stage != release {
		v.stage = release
	}
	switch v.stage {
	case attack:
		if in.Attack <= 0 {
			v.env = 1
		} else {
			v.env += dt / in.Attack
		}
		if v.env >= 1 {
			v.env, v.stage = 1, decay
		}
	case decay:
		if in.Decay <= 0 {
			v.env = in.Sustain
		} else {
			v.env -= dt / in.Decay * (1 - in.Sustain)
		}
		if v.env <= in.Sustain {
			v.env, v.stage = in.Sustain, sustain
		}
		if v.env <= 0 {
			v.stage = idle
		}
	case release:
		if in.Release <= 0 {
			v.env = 0
		} else {
			v.env -= dt / in.Release
		}
		if v.env <= 0 {
			v.env, v.stage = 0, idle
		}
	}
}

// clockNoise advances the NES's 15-bit LFSR at rate clocks per sample,
// at most one, and returns its output bit as ±1.
func (v *voice) clockNoise(rate float64) float64 {
	v.noise += min(1, rate)
	if v.noise >= 1 {
		v.noise--
		bit := (v.lfsr ^ v.lfsr>>1) & 1
		v.lfsr = v.lfsr>>1 | bit<<14
	}
	if v.lfsr&1 == 0 {
		return 1
	}
	return -1
}

// pulse is a band-limited pulse wave at phase with duty cycle duty.
func pulse(phase, inc, duty float64) float64 {
	s := -1.0
	if phase < duty {
		s = 1
	}
	fall := phase - duty
	if fall < 0 {
		fall++
	}
	return s + blep(phase, inc) - blep(fall, inc)
}

// triangle is the NES triangle: 32 steps of a 4-bit ramp up and down.
func triangle(phase float64) float64 {
	step := int(phase * 32)
	level := step - 16
	if step < 16 {
		level = 15 - step
	}
	return float64(level)/7.5 - 1
}

// blep is the PolyBLEP correction for a step at phase 0, which keeps
// pulses and saws from aliasing into a mess at 22 kHz.
func blep(t, dt float64) float64 {
	switch {
	case t < dt:
		t /= dt
		return t + t - t*t - 1
	case t > 1-dt:
		t = (t - 1) / dt
		return t*t + t + t + 1
	}
	return 0
}
