// Package music implements gibson music, which plays the soundtrack as a
// WAV on stdout, or writes a stretch of it to a file.
package music

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/spechtlabs/gibson/internal/music"
	"github.com/spechtlabs/gibson/pkg/chip"
)

// ErrTerminal is returned when the stream would go to a terminal.
var ErrTerminal = errors.New("stdout is a terminal: pipe it into a player, e.g. gibson music | ffplay -nodisp -")

// NewCommand returns the music command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "music",
		Short: "Play the soundtrack as a WAV on stdout",
		Long: `Streams the soundtrack, an 8-bit chiptune, as a WAV on stdout in real time,
for a player to read from a pipe; gibson serve streams the same to
ssh host music. With --out, it writes --dur of it to a file instead.`,
		Example: `# Listen while the intro plays
gibson music | ffplay -nodisp -autoexit -loglevel quiet - & gibson

# A minute of it to a file
gibson music --out gibson.wav --dur 1m`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out, _ := cmd.Flags().GetString("out")
			if out != "" {
				dur, _ := cmd.Flags().GetDuration("dur")
				return write(out, dur)
			}
			if term.IsTerminal(os.Stdout.Fd()) {
				return ErrTerminal
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			if err := music.Play(ctx, os.Stdout); err != nil && ctx.Err() == nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().String("out", "", "write the soundtrack to this WAV file instead")
	cmd.Flags().Duration("dur", time.Minute, "how much of it --out gets")
	return cmd
}

func write(path string, dur time.Duration) error {
	s, err := music.Song()
	if err != nil {
		return err
	}
	p, err := chip.NewPlayer(s, music.Rate)
	if err != nil {
		return err
	}
	n := int(dur.Seconds() * music.Rate)
	f, err := os.Create(path) //nolint:gosec // the file the user asked for
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	_, _ = w.Write(chip.WAVHeader(music.Rate, n))
	if _, err := io.CopyN(w, p, int64(n)); err != nil {
		_ = f.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	return f.Close()
}
