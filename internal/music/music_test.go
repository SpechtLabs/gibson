package music

import (
	"math"
	"testing"

	"github.com/spechtlabs/gibson/pkg/chip"
)

func TestSongPlays(t *testing.T) {
	s, err := Song()
	if err != nil {
		t.Fatal(err)
	}
	p, err := chip.NewPlayer(s, Rate)
	if err != nil {
		t.Fatal(err)
	}
	// one pass through the song and back into the loop
	buf := make([]byte, int(float64(s.Rows()+bar)*s.RowSeconds()*Rate))
	_, _ = p.Read(buf)

	var sum float64
	var clipped int
	for _, b := range buf {
		x := (float64(b) - 127.5) / 127.5
		sum += x * x
		if b == 0 || b == 255 {
			clipped++
		}
	}
	rms := math.Sqrt(sum / float64(len(buf)))
	t.Logf("%.1f s, rms %.3f, clipped %.3f%%", float64(len(buf))/Rate, rms, 100*float64(clipped)/float64(len(buf)))
	if rms < 0.05 {
		t.Errorf("rms %.3f: the song is close to silent", rms)
	}
	if frac := float64(clipped) / float64(len(buf)); frac > 0.001 {
		t.Errorf("%.2f%% of samples clip", 100*frac)
	}
}

func TestDropLandsOnTheTitle(t *testing.T) {
	// the title starts decrypting at 4.7 s in internal/scene
	s, err := Song()
	if err != nil {
		t.Fatal(err)
	}
	if at := float64(introOf) * s.RowSeconds(); math.Abs(at-4.8) > 0.15 {
		t.Errorf("the drop lands at %.2f s, not with the title", at)
	}
}
