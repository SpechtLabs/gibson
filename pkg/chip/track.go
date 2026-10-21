package chip

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Row is one step of a track: a note to play, a note-off, or nothing.
type Row struct {
	Note int         // MIDI note number (60 is C4), or [Hold] or [Off]
	Inst *Instrument // what plays Note
	Arp  []int       // semitone offsets cycled once per tick, tracker style
	Vol  float64     // velocity, 0 to 1
}

// What a [Row] does besides playing a note.
const (
	Hold = 0  // let the channel's note ring on
	Off  = -1 // release the channel's note
)

// Track is a channel's rows, one after another.
type Track []Row

// Kit names the instruments a track can play. The instrument under ""
// plays every note that doesn't name one.
type Kit map[string]*Instrument

// Parse reads a track written tracker style, one whitespace-separated
// token per row:
//
//	.         nothing: the note rings on
//	-         note off
//	A4 C#5 Bb3  a note, played by kit[""]
//	A3/037c   a note arpeggiated through the hex offsets 0, 3, 7 and 12
//	E5:7      a note at velocity (7+1)/16, hex 0 to f
//	A6@b      a note played by kit["b"]
//	k         kit["k"] at its own note
//	k:5       either of the last two at a velocity
//
// In full, a note is NOTE[/ARP][@KIT][:VEL].
func Parse(kit Kit, s string) (Track, error) {
	fields := strings.Fields(s)
	t := make(Track, 0, len(fields))
	for _, tok := range fields {
		r, err := parseRow(kit, tok)
		if err != nil {
			return nil, fmt.Errorf("row %d %q: %w", len(t), tok, err)
		}
		t = append(t, r)
	}
	return t, nil
}

// Parser parses track after track and keeps the first error, so a song
// written in code checks once, at the end, rather than after every track.
type Parser struct{ err error }

// Parse is [Parse], returning an empty track once anything has failed.
func (p *Parser) Parse(kit Kit, s string) Track {
	if p.err != nil {
		return nil
	}
	t, err := Parse(kit, s)
	if err != nil {
		p.err = fmt.Errorf("chip: %w", err)
	}
	return t
}

// Err is the first error any track had.
func (p *Parser) Err() error { return p.err }

func parseRow(kit Kit, tok string) (Row, error) {
	switch tok {
	case ".":
		return Row{Note: Hold}, nil
	case "-":
		return Row{Note: Off}, nil
	}
	r := Row{Vol: 1, Inst: kit[""]}
	if body, vol, ok := strings.Cut(tok, ":"); ok {
		v, err := strconv.ParseUint(vol, 16, 4)
		if err != nil {
			return Row{}, fmt.Errorf("velocity %q is not one hex digit", vol)
		}
		r.Vol, tok = float64(v+1)/16, body
	}
	if inst, ok := kit[tok]; ok && tok != "" {
		r.Note, r.Inst = inst.Note, inst
		return r, nil
	}
	if body, name, ok := strings.Cut(tok, "@"); ok {
		inst, found := kit[name]
		if !found || name == "" {
			return Row{}, fmt.Errorf("no instrument %q in the kit", name)
		}
		r.Inst, tok = inst, body
	}
	if body, arp, ok := strings.Cut(tok, "/"); ok {
		for _, c := range arp {
			v, err := strconv.ParseUint(string(c), 16, 4)
			if err != nil {
				return Row{}, fmt.Errorf("arpeggio offset %q is not a hex digit", c)
			}
			r.Arp = append(r.Arp, int(v))
		}
		tok = body
	}
	n, err := ParseNote(tok)
	if err != nil {
		return Row{}, err
	}
	if r.Inst == nil {
		return Row{}, fmt.Errorf("no instrument for %q: the kit has none under \"\"", tok)
	}
	r.Note = n
	return r, nil
}

// ErrNoNote is returned for an empty note name.
var ErrNoNote = errors.New("no note")

var semitones = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

// ParseNote turns a note name such as A4, C#5 or Bb3 into its MIDI number.
func ParseNote(s string) (int, error) {
	if s == "" {
		return 0, ErrNoNote
	}
	n, ok := semitones[s[0]]
	if !ok {
		return 0, fmt.Errorf("%q is not a note", s)
	}
	rest := s[1:]
	switch {
	case strings.HasPrefix(rest, "#"):
		n, rest = n+1, rest[1:]
	case strings.HasPrefix(rest, "b"):
		n, rest = n-1, rest[1:]
	}
	oct, err := strconv.Atoi(rest)
	if err != nil || oct < 0 || oct > 9 {
		return 0, fmt.Errorf("%q has no octave from 0 to 9", s)
	}
	return 12*(oct+1) + n, nil
}

// Transpose returns t with every note shifted by n semitones.
func (t Track) Transpose(n int) Track {
	out := make(Track, len(t))
	for i, r := range t {
		if r.Note > 0 {
			r.Note += n
		}
		out[i] = r
	}
	return out
}

// Stretch returns t with n-1 holds after every row, so a track written in
// eighth notes plays on a sixteenth-note grid with n = 2.
func (t Track) Stretch(n int) Track {
	out := make(Track, 0, len(t)*n)
	for _, r := range t {
		out = append(out, r)
		for range n - 1 {
			out = append(out, Row{Note: Hold})
		}
	}
	return out
}

// Times returns t played n times over.
func (t Track) Times(n int) Track {
	out := make(Track, 0, len(t)*n)
	for range n {
		out = append(out, t...)
	}
	return out
}

// Cat joins tracks end to end.
func Cat(ts ...Track) Track {
	var out Track
	for _, t := range ts {
		out = append(out, t...)
	}
	return out
}

// Rest is n rows of nothing.
func Rest(n int) Track { return make(Track, n) }
