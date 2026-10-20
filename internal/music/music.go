// Package music is the gibson's soundtrack: a chiptune played by
// pkg/chip, and the SSH handler that streams it. SSH carries bytes, not
// sound, so the stream is an 8-bit WAV on stdout for the viewer to pipe
// into a player next to the intro:
//
//	ssh -p 2222 host music | ffplay -nodisp -autoexit -loglevel quiet -
package music

import (
	"context"
	"fmt"
	"io"
	"time"

	"charm.land/ssh"
	"charm.land/wish/v2"

	"github.com/spechtlabs/gibson/pkg/chip"
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

// Handler streams the music to an SSH session. It needs a session
// without a pty: a pty would turn every newline byte of the audio into
// CRLF.
func Handler(sess ssh.Session) {
	if _, _, ok := sess.Pty(); ok {
		wish.Fatalln(sess, "music is an audio stream: leave out -t and pipe it into a player, e.g.\n"+
			"  ssh <host> music | ffplay -nodisp -autoexit -loglevel quiet -")
		return
	}
	_, _ = fmt.Fprintf(sess.Stderr(), "♪ now playing: %q, %d Hz 8-bit mono, looping\n", Title, Rate)
	_ = Play(sess.Context(), sess)
}
