// Package denied implements gibson denied, which plays the screen a viewer
// gets when gibson serve is at capacity, right here in the terminal.
package denied

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/app"
	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/denied"
)

// NewCommand returns the denied command, which reads the look from v.
func NewCommand(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denied",
		Short: "Preview the screen a viewer gets when the server is full",
		Long: `Plays the screen gibson serve shows a viewer it turns away at
--max-sessions: the scene frozen, a scan line sweeping down it, and a red
ACCESS DENIED with a line from Hackers (1995). The look flags pick the scene,
as session words do over SSH.

--quote picks the line by number; by default it's random, as over SSH:
` + list(),
		Example: `gibson denied --theme specht --vessel dodeca
gibson denied --quote 0`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			look, err := config.Look(v)
			if err != nil {
				return err
			}
			sc, err := look.New()
			if err != nil {
				return err
			}
			quote := denied.Quote()
			if n, _ := cmd.Flags().GetInt("quote"); n >= 0 {
				if n >= len(denied.Quotes) {
					return fmt.Errorf("--quote %d: there are %d quotes, 0 to %d", n, len(denied.Quotes), len(denied.Quotes)-1)
				}
				quote = denied.Quotes[n]
			}
			w, h, err := term.GetSize(os.Stdout.Fd())
			if err != nil {
				w, h = 120, 40 // not a terminal: still produce something
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return denied.Play(ctx, os.Stdout, sc, quote, min(w, app.MaxCols), min(h, app.MaxRows))
		},
	}
	cmd.Flags().Int("quote", -1, "which line to show, by number (default random)")
	return cmd
}

func list() string {
	var b strings.Builder
	for i, q := range denied.Quotes {
		fmt.Fprintf(&b, "\n  %d  %s", i, q)
	}
	return b.String()
}
