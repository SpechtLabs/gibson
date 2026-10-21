// Package music is the gibson's soundtrack: a chiptune played by
// pkg/chip, the SSH handler that streams it, and playback on this
// machine's speakers. SSH carries bytes, not sound, so the stream is an
// 8-bit WAV, for gibson connect to play, or a player: on stdout alone, or
// on stderr beside the intro on the terminal (fish's 2>|, bash's 2> >()):
//
//	ssh -p 2222 host music | ffplay -nodisp -autoexit -loglevel quiet -
//	ssh -t -p 2222 host music 2>| ffplay -nodisp -autoexit -loglevel quiet -
package music

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"charm.land/ssh"

	"github.com/spechtlabs/gibson/pkg/chip"
	"github.com/spechtlabs/gibson/pkg/speaker"
)

// Rate is the sample rate: 22.05 kHz, 8-bit mono, about 22 KB/s, a
// twentieth of what the picture costs.
const Rate = 22050

// ahead is how much is sent ahead of real time, enough to ride out a
// stall on the way without starting far ahead of the picture.
const ahead = 300 * time.Millisecond

// Command is the SSH command that streams the music.
const Command = "music"

// Play streams the song to w in real time until ctx is done or w fails.
func Play(ctx context.Context, w io.Writer) error {
	s, err := Song()
	if err != nil {
		return err
	}
	p, err := chip.NewPlayer(s, Rate)
	if err != nil {
		return err
	}
	return chip.Stream(ctx, w, p, Rate, ahead)
}

// Handler streams the music to an SSH session without a pty.
func Handler(sess ssh.Session) {
	_, _ = fmt.Fprintf(sess.Stderr(), "♪ now playing: %q, %d Hz 8-bit mono, looping\n", Title, Rate)
	_ = Play(sess.Context(), sess)
}

// Alongside streams the music over sess's stderr, beside the intro on its
// pty, and returns what starts the song over with the intro. A pty
// session's stderr turns every newline byte into CRLF, so the stream has
// none: the header has none at Rate, and samples of 10 go out as 11, a
// step no one hears.
func Alongside(sess ssh.Session) (restart func()) {
	src, err := NewSource(0)
	if err != nil { // the song is fixed, and tested to play
		return func() {}
	}
	go func() { _ = chip.Stream(sess.Context(), noNewlines{sess.Stderr()}, src, Rate, ahead) }()
	return src.Restart
}

// noNewlines writes to w with every newline byte one higher.
type noNewlines struct{ w io.Writer }

func (n noNewlines) Write(b []byte) (int, error) {
	return n.w.Write(bytes.ReplaceAll(b, []byte{'\n'}, []byte{'\n' + 1}))
}

// Source is the song as PCM for the speakers, which can start over.
type Source struct {
	mu   sync.Mutex
	song *chip.Song
	p    *chip.Player
}

// NewSource returns the song from offset seconds in.
func NewSource(offset float64) (*Source, error) {
	s, err := Song()
	if err != nil {
		return nil, err
	}
	p, err := playFrom(s, offset)
	if err != nil {
		return nil, err
	}
	return &Source{song: s, p: p}, nil
}

// Restart starts the song over, as the intro does on r.
func (src *Source) Restart() {
	src.mu.Lock()
	defer src.mu.Unlock()
	if p, err := playFrom(src.song, 0); err == nil { // it played once, so it can't fail
		src.p = p
	}
}

// playFrom returns a player of s, offset seconds in.
func playFrom(s *chip.Song, offset float64) (*chip.Player, error) {
	p, err := chip.NewPlayer(s, Rate)
	if err != nil {
		return nil, err
	}
	if _, err := io.CopyN(io.Discard, p, int64(offset*Rate)); err != nil {
		return nil, err
	}
	return p, nil
}

// Read fills b with the next samples; the song never ends.
func (src *Source) Read(b []byte) (int, error) {
	src.mu.Lock()
	defer src.mu.Unlock()
	return src.p.Read(b)
}

// Listen plays a stream from [Handler], as ssh host music sends it, on
// the speakers until it ends or ctx is done.
func Listen(ctx context.Context, r io.Reader) error {
	rate, err := chip.ReadWAVHeader(r)
	if err != nil {
		return err
	}
	return speaker.Play(ctx, r, rate)
}
