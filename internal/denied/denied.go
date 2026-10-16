// Package denied is what a viewer gets when the gibson is full: the scene
// frozen in its held moment, one scan line sweeping down it, and then a red
// ACCESS DENIED with a movie line, Hackers (1995) to begin with, before the
// session closes.
//
// It's built to be cheap, since it only runs when the server is busy: one
// frame is rendered, and the sweep redraws only the rows it passes.
package denied

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/ssh"
	"charm.land/wish/v2"

	"github.com/spechtlabs/gibson/internal/scene"
	"github.com/spechtlabs/gibson/pkg/components/banner"
	r "github.com/spechtlabs/gibson/pkg/render"
)

const (
	heading = "ACCESS DENIED"

	// holdAt is the moment the scene freezes at: everything is up, the
	// title's sheen is off to the side and no glitch is tearing it.
	holdAt = 9.8
	// sweep is how long the scan line takes from top to bottom.
	sweep = 900 * time.Millisecond
)

// Handler returns the capacity handler for a server whose sessions look
// like look, with terminals capped at maxCols×maxRows.
func Handler(look scene.Look, maxCols, maxRows int) ssh.Handler {
	return func(sess ssh.Session) {
		quote := Pick()
		pty, _, ok := sess.Pty()
		w, h := min(pty.Window.Width, maxCols), min(pty.Window.Height, maxRows)
		// session words are all known names, so this can't fail
		sc, err := look.Override(sess.Command()).New()
		if !ok || w < 20 || h < 8 || err != nil {
			wish.Fatalln(sess, heading+": "+quote.Text+" "+quote.Attribution())
			return
		}
		_ = Play(sess.Context(), sess, sc, quote, w, h)
		_ = sess.Exit(1)
	}
}

// Play plays the denial for sc on a w×h terminal behind out, and leaves it
// on the screen. It stops early, with ctx's error, when ctx is done, and
// with the write error when out stops taking writes.
func Play(ctx context.Context, out io.Writer, sc *scene.Scene, quote Quote, w, h int) error {
	frame := sc.Render(holdAt, w, h).Clone()
	glow := sc.Palette().Grad[1].Lerp(r.RGB{R: 1, G: 1, B: 1}, 0.3)

	var b strings.Builder
	b.WriteString("\x1b[?25l\x1b[2J")
	for y := range h {
		row(&b, frame, y)
	}
	if err := write(out, &b); err != nil {
		return err
	}

	// the scan line: a bright row with a fading trail, restoring the rows
	// behind it as it goes
	trail := []float64{1, 0.45, 0.15}
	step := sweep / time.Duration(h)
	for y := 0; y < h+len(trail); y++ {
		for k, s := range trail {
			if yy := y - k; yy >= 0 && yy < h {
				lit(&b, frame, yy, glow, s)
			}
		}
		if yy := y - len(trail); yy >= 0 && yy < h {
			row(&b, frame, yy)
		}
		if err := write(out, &b); err != nil {
			return err
		}
		if err := pause(ctx, step); err != nil {
			return err
		}
	}

	denied := frame.Clone()
	banner.New(heading, quote.Text, banner.WithAttribution(quote.Attribution())).Apply(denied)
	for y := range h {
		row(&b, denied, y)
	}
	// leave the cursor under it, so the shell prompt comes back below
	fmt.Fprintf(&b, "\x1b[0m\x1b[%d;1H\r\n\x1b[?25h", h)
	return write(out, &b)
}

// row moves to row y and draws it.
func row(b *strings.Builder, f *r.Frame, y int) {
	fmt.Fprintf(b, "\x1b[%d;1H", y+1)
	b.WriteString(f.RowANSI(y))
}

// lit draws row y with the scan line's glow on it at strength s.
func lit(b *strings.Builder, f *r.Frame, y int, glow r.RGB, s float64) {
	band := &r.Frame{W: f.W, H: 1, Cells: append([]r.Cell(nil), f.Cells[y*f.W:(y+1)*f.W]...)}
	for i := range band.Cells {
		c := &band.Cells[i]
		c.BG = c.BG.Add(glow.Scale(0.35 * s))
		c.FG = c.FG.Lerp(r.RGB{R: 1, G: 1, B: 1}, 0.4*s)
	}
	fmt.Fprintf(b, "\x1b[%d;1H", y+1)
	b.WriteString(band.RowANSI(0))
}

// write sends what b holds to out and empties b.
func write(out io.Writer, b *strings.Builder) error {
	_, err := io.WriteString(out, b.String())
	b.Reset()
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// pause waits d, or until ctx is done.
func pause(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
