package music

import (
	"strings"

	"github.com/spechtlabs/gibson/pkg/chip"
)

// The instruments, C64 and NES flavored.
var (
	lead = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.5,
		Duty: 0.25, PWMDepth: 0.08, PWMRate: 0.7,
		Attack: 0.004, Decay: 0.3, Sustain: 0.6, Release: 0.1,
		Vibrato: 0.22, VibratoRate: 5.5, VibratoDelay: 0.2,
		Glide: 0.035,
	}
	arp = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.28,
		// the slow pulse-width sweep is the SID's signature
		Duty: 0.5, PWMDepth: 0.38, PWMRate: 0.23,
		Decay: 0.16, Sustain: 0.4, Release: 0.05,
	}
	bass = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.5,
		Duty: 0.3, PWMDepth: 0.12, PWMRate: 0.4,
		Decay: 0.12, Sustain: 0.55, Release: 0.04,
	}
	pad = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.24,
		Duty: 0.5, PWMDepth: 0.3, PWMRate: 0.3,
		Attack: 0.5, Sustain: 1, Release: 1,
		Vibrato: 0.1, VibratoRate: 4,
	}
	sparkle = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.14,
		Duty: 0.125, Decay: 0.09,
	}
	blip = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.18,
		Duty: 0.5, Decay: 0.05,
	}
	// the CRT snapping on
	zap = &chip.Instrument{
		Wave: chip.Pulse, Volume: 0.35, Note: 96,
		Duty: 0.5, Decay: 0.35, Sweep: -150, Floor: 30,
	}

	kick = &chip.Instrument{
		Wave: chip.Triangle, Volume: 1, Note: 52,
		Decay: 0.2, Sweep: -200, Floor: 28,
	}
	snare = &chip.Instrument{
		Wave: chip.Noise, Volume: 0.55, Note: 112,
		Decay: 0.17, Sweep: -45,
	}
	hat = &chip.Instrument{
		Wave: chip.Noise, Volume: 0.22, Note: 130,
		Decay: 0.035,
	}
	openHat = &chip.Instrument{
		Wave: chip.Noise, Volume: 0.18, Note: 128,
		Decay: 0.2,
	}
	crash = &chip.Instrument{
		Wave: chip.Noise, Volume: 0.4, Note: 124,
		Decay: 1.4,
	}
	// white noise swelling and rising into the drop
	riser = &chip.Instrument{
		Wave: chip.Noise, Volume: 0.3, Note: 96,
		Attack: 3.2, Sustain: 1, Sweep: 9,
	}
)

var (
	leads   = chip.Kit{"": lead}
	arps    = chip.Kit{"": arp}
	basses  = chip.Kit{"": bass}
	fx      = chip.Kit{"": pad, "z": zap}
	blips   = chip.Kit{"": blip}
	sparkls = chip.Kit{"": sparkle}
	drums   = chip.Kit{"k": kick, "s": snare}
	hats    = chip.Kit{"h": hat, "o": openHat, "x": crash, "r": riser}
)

// The song runs at speed 5, ten rows a second: 150 BPM in sixteenths, a
// bar every 1.6 s. The intro is three bars, so the drop lands at 4.8 s,
// as the title starts decrypting; the boot log gets blips, the CRT
// snapping on gets a zap and the materialization gets the riser.
const (
	speed   = 5
	bar     = 16
	introOf = 3 * bar
)

// Title is the song's name.
const Title = "Containment Breach"

// The chords, as bass transpositions from A and as arpeggios voiced
// around A3, so the arpeggio doesn't jump about.
const (
	am, f, c, g, em, e = 0, -4, 3, -2, -5, -5

	aAm, aF, aC, aG = "A3/037c", "A3/038c", "G3/059c", "G3/047c"
	aEm, aE         = "G3/049c", "G#3/038c"
)

// section is a stretch of the song, one track per channel, in the order
// of channels.
type section [6]chip.Track

var channels = [...]struct {
	name       string
	gain, echo float64
}{
	{"lead", 0.55, 0.45},
	{"arp", 0.5, 0.15},
	{"bass", 0.55, 0},
	{"drums", 0.55, 0},
	{"hats", 0.5, 0},
	{"fx", 0.5, 0.5},
}

// score writes the song's parts with the shorthands they share.
type score struct{ chip.Parser }

// sixteenths reads s a row a token.
func (sc *score) sixteenths(kit chip.Kit, s string) chip.Track { return sc.Parse(kit, s) }

// eighths reads s in eighth notes onto the sixteenth-note grid.
func (sc *score) eighths(kit chip.Kit, s string) chip.Track { return sc.Parse(kit, s).Stretch(2) }

// chords is a bar of each arpeggio, retriggered every eighth.
func (sc *score) chords(toks ...string) chip.Track {
	t := make(chip.Track, 0, len(toks)*bar)
	for _, tok := range toks {
		t = append(t, sc.eighths(arps, strings.Repeat(tok+" ", 8))...)
	}
	return t
}

// roots is a bar of octave-bouncing bass, synthwave style, on each root.
func (sc *score) roots(semis ...int) chip.Track {
	octave := sc.eighths(basses, "A2 A3 A2 A3 A2 A3 A2 A3")
	t := make(chip.Track, 0, len(semis)*bar)
	for _, s := range semis {
		t = append(t, octave.Transpose(s)...)
	}
	return t
}

func (sc *score) fill() chip.Track {
	return sc.sixteenths(drums, "k . . . s . . . k . s . s s s s")
}

// groove is eight bars of beat, the last a fill.
func (sc *score) groove() chip.Track {
	beat := sc.sixteenths(drums, "k . . . s . . . k . . k s . . .")
	return chip.Cat(beat.Times(7), sc.fill())
}

// hatLine is eight bars of hats, opening on a crash.
func (sc *score) hatLine() chip.Track {
	return chip.Cat(
		sc.sixteenths(hats, "x . o . h . o . h . o . h . o ."),
		sc.sixteenths(hats, "h . o . h . o . h . o . h h o .").Times(7),
	)
}

// pickup leads into the hook from the end of a bar.
func (sc *score) pickup() chip.Track {
	return sc.sixteenths(leads, ". . . . . . . . . . . . E4 . G4 .")
}

// intro plays under the boot: blips for the boot log, the CRT's zap, a
// drone and an arpeggio fading in as the vessel builds, and the bass, a
// riser and a snare roll under the materialization.
func (sc *score) intro() section {
	return section{
		chip.Cat(
			sc.sixteenths(blips, ". . . . . A6 . E6 . . A6 . C7 . . E6"),
			sc.sixteenths(blips, ". A6 . . E6 . G6 . . A6 . . . E7 . ."),
			sc.pickup(),
		),
		chip.Cat(chip.Rest(bar), sc.eighths(arps, `
			A3/037c:2 A3/037c:3 A3/037c:4 A3/037c:5 A3/037c:6 A3/037c:7 A3/037c:8 A3/037c:9
			A3/037c:a A3/037c:a A3/037c:b A3/037c:b A3/037c:c A3/037c:d A3/037c:e A3/037c:f`)),
		chip.Cat(chip.Rest(2*bar), sc.eighths(basses, "A2:7 A2:8 A2:9 A2:a A2:b A2:c A2:d A2:f")),
		chip.Cat(chip.Rest(2*bar), sc.fill()),
		chip.Cat(chip.Rest(bar), sc.sixteenths(hats, "r"), chip.Rest(2*bar-1)),
		chip.Cat(sc.sixteenths(fx, ". . z . . . . . . . . . . . . A3"), chip.Rest(2*bar)),
	}
}

// partA is the hook over Am F C G.
func (sc *score) partA() section {
	return section{
		sc.eighths(leads, `
			A4 C5 E5 .  .  D5 C5 D5
			C5 .  .  A4 .  .  .  -
			G4 C5 E5 .  .  D5 E5 G5
			.  .  .  D5 .  .  B4 -
			A5 .  G5 E5 .  .  C5 E5
			F5 .  E5 C5 .  .  A4 C5
			E5 .  D5 C5 .  G4 C5 E5
			D5 .  .  .  B4 .  D5 -`),
		sc.chords(aAm, aF, aC, aG, aAm, aF, aC, aG),
		sc.roots(am, f, c, g, am, f, c, g),
		sc.groove(),
		sc.hatLine(),
		chip.Cat(sc.sixteenths(fx, "-"), chip.Rest(8*bar-1)),
	}
}

// partB climbs, and turns to E major to pull back to the hook.
func (sc *score) partB() section {
	return section{
		sc.eighths(leads, `
			A5 .  .  .  G5 .  F5 .
			G5 .  .  .  .  .  D5 .
			E5 .  .  .  G5 .  B5 .
			A5 .  .  .  .  .  .  -
			C6 .  B5 .  A5 .  G5 .
			B5 .  A5 .  G5 .  D5 .
			E5 .  G#5 . B5 .  .  .
			G#5 . .  .  .  .  .  -`),
		sc.chords(aF, aG, aEm, aAm, aF, aG, aE, aE),
		sc.roots(f, g, em, am, f, g, e, e),
		sc.groove(),
		sc.hatLine(),
		chip.Rest(8 * bar),
	}
}

// breakdown drops the beat, the pad sings over the changes, and a roll
// and the riser bring the hook back.
func (sc *score) breakdown() section {
	return section{
		chip.Cat(chip.Rest(3*bar), sc.pickup()),
		sc.chords(aAm, aF, aC, aG),
		sc.eighths(basses, `
			A2 . . . . . . .
			F2 . . . . . . .
			C3 . . . . . . .
			G2 . . . . . . -`),
		chip.Cat(sc.sixteenths(drums, "k . . . . . . . . . . . . . . .").Times(3), sc.fill()),
		chip.Cat(
			sc.sixteenths(hats, ". . h:5 . . . h:5 . . . h:5 . . . h:5 .").Times(2),
			sc.sixteenths(hats, "r"), chip.Rest(2*bar-1),
		),
		sc.eighths(fx, `
			E4 . . . . . . .
			F4 . . . . . . .
			E4 . . . . . . .
			D4 . . . . . . -`),
	}
}

// partC is the hook again, with a demoscene sparkle on top.
func (sc *score) partC() section {
	s := sc.partA()
	s[5] = sc.sixteenths(sparkls, `
		A5 C6 E6 C6 A5 C6 E6 C6 A5 C6 E6 C6 A5 C6 E6 C6
		A5 C6 F6 C6 A5 C6 F6 C6 A5 C6 F6 C6 A5 C6 F6 C6
		G5 C6 E6 C6 G5 C6 E6 C6 G5 C6 E6 C6 G5 C6 E6 C6
		G5 B5 D6 B5 G5 B5 D6 B5 G5 B5 D6 B5 G5 B5 D6 B5`).Times(2)
	return s
}

// Song is the gibson's theme, [Title]: A minor, synthwave changes, a
// demoscene arpeggio channel and a hook with echo on it. It plays the
// intro once and then loops the rest.
func Song() (*chip.Song, error) {
	var sc score
	parts := []section{sc.intro(), sc.partA(), sc.partB(), sc.breakdown(), sc.partC()}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	song := &chip.Song{
		Title:        Title,
		Speed:        speed,
		Loop:         introOf,
		EchoRows:     3, // a dotted eighth
		EchoFeedback: 0.35,
	}
	for i, ch := range channels {
		var t chip.Track
		for _, p := range parts {
			t = append(t, p[i]...)
		}
		song.Channels = append(song.Channels, chip.Channel{Name: ch.name, Track: t, Gain: ch.gain, Echo: ch.echo})
	}
	return song, song.Validate()
}
