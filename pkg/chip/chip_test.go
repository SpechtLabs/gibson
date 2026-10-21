package chip

import (
	"bytes"
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	lead := &Instrument{Wave: Pulse, Volume: 1}
	hat := &Instrument{Wave: Noise, Volume: 1, Note: 130}
	kit := Kit{"": lead, "h": hat}
	tr, err := Parse(kit, ". - A4 C#5 Bb3 A3/037c E5:7 A6@h h h:3")
	if err != nil {
		t.Fatal(err)
	}
	notes := make([]int, len(tr))
	for i, r := range tr {
		notes[i] = r.Note
	}
	if want := []int{Hold, Off, 69, 73, 58, 57, 76, 93, 130, 130}; !slices.Equal(notes, want) {
		t.Errorf("notes %v, want %v", notes, want)
	}
	if !slices.Equal(tr[5].Arp, []int{0, 3, 7, 12}) {
		t.Errorf("arp %v", tr[5].Arp)
	}
	if tr[6].Vol != 0.5 || tr[9].Vol != 0.25 {
		t.Errorf("velocities %v and %v", tr[6].Vol, tr[9].Vol)
	}
	if tr[7].Inst != hat || tr[8].Inst != hat || tr[2].Inst != lead {
		t.Error("wrong instruments")
	}
	for _, bad := range []string{"H4", "A", "A4@x", "A4:g", "A4/z", "k"} {
		if _, err := Parse(Kit{"": lead}, bad); err == nil {
			t.Errorf("Parse(%q) didn't fail", bad)
		}
	}
}

func TestTrackHelpers(t *testing.T) {
	tr := mustParse(t, Kit{"": {}}, "A4 -")
	if got := tr.Stretch(2); len(got) != 4 || got[0].Note != 69 || got[1].Note != Hold || got[2].Note != Off {
		t.Errorf("Stretch: %v", got)
	}
	if got := tr.Transpose(3); got[0].Note != 72 || got[1].Note != Off || tr[0].Note != 69 {
		t.Errorf("Transpose: %v", got)
	}
	if got := Cat(tr, Rest(2)).Times(2); len(got) != 8 {
		t.Errorf("Cat/Times: %d rows", len(got))
	}
}

func TestPlayerIsDeterministic(t *testing.T) {
	inst := &Instrument{Wave: Pulse, Volume: 1, Duty: 0.5, Sustain: 1}
	s := &Song{Speed: 6, Channels: []Channel{{Name: "a", Gain: 1, Track: mustParse(t, Kit{"": inst}, "A4/047 . - .")}}}
	a, b := make([]byte, 22050), make([]byte, 22050)
	pa, err := NewPlayer(s, 22050)
	if err != nil {
		t.Fatal(err)
	}
	pb, _ := NewPlayer(s, 22050)
	_, _ = pa.Read(a)
	_, _ = pb.Read(b)
	if !bytes.Equal(a, b) {
		t.Error("two players of one song differ")
	}
	if bytes.Count(a, []byte{128}) == len(a) {
		t.Error("silent")
	}
}

func TestValidate(t *testing.T) {
	k := Kit{"": {}}
	for name, s := range map[string]*Song{
		"empty":   {Speed: 6},
		"ragged":  {Speed: 6, Channels: []Channel{{Track: mustParse(t, k, "A4 .")}, {Track: mustParse(t, k, "A4")}}},
		"speed 0": {Channels: []Channel{{Track: mustParse(t, k, "A4")}}},
		"loop":    {Speed: 6, Loop: 1, Channels: []Channel{{Track: mustParse(t, k, "A4")}}},
	} {
		if s.Validate() == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestWAVHeader(t *testing.T) {
	h := WAVHeader(22050, 100)
	if len(h) != 44 || string(h[:4]) != "RIFF" || string(h[36:40]) != "data" || h[40] != 100 || h[4] != 136 {
		t.Errorf("header % x", h)
	}
	if s := WAVHeader(22050, Streaming); !bytes.Equal(s[40:], []byte{0xff, 0xff, 0xff, 0xff}) {
		t.Errorf("streaming header % x", s)
	}
}

func mustParse(t *testing.T, kit Kit, s string) Track {
	t.Helper()
	tr, err := Parse(kit, s)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func TestReadWAVHeader(t *testing.T) {
	r := bytes.NewReader(append(WAVHeader(22050, Streaming), 1, 2, 3))
	rate, err := ReadWAVHeader(r)
	if err != nil || rate != 22050 {
		t.Fatalf("rate %d, err %v", rate, err)
	}
	if n := r.Len(); n != 3 {
		t.Errorf("%d bytes left after the header, want the 3 samples", n)
	}
	h := WAVHeader(22050, 0)
	h[34] = 16 // 16-bit
	if _, err := ReadWAVHeader(bytes.NewReader(h)); err == nil {
		t.Error("took a 16-bit WAV")
	}
}
