// Package connect implements gibson connect, which watches a gibson served
// over SSH with its soundtrack playing on this machine's speakers.
package connect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/music"
)

// NewCommand returns the connect command, which reads --music from v.
func NewCommand(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [user@]host [look words...]",
		Short: "Watch a gibson over SSH, with sound",
		Long: `Connects to a gibson served over SSH and plays its soundtrack on this
machine's speakers while the intro runs, which plain ssh can't: SSH carries
bytes, not sound.

It runs your ssh twice, once for the intro and once for the music, so your
ssh config, keys and known_hosts apply as they always do. Words after the
host pick the look, as they do over ssh.`,
		Example: `gibson connect gibson.example.com
gibson connect -p 2222 gibson.example.com dodeca specht`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Bind(v, cmd.Flags(), map[string]string{"music": config.Music}); err != nil {
				return err
			}
			ssh, err := exec.LookPath("ssh")
			if err != nil {
				return fmt.Errorf("gibson connect runs ssh: %w", err)
			}
			port, _ := cmd.Flags().GetInt("port")
			return watch(cmd.Context(), ssh, strconv.Itoa(port), args[0], args[1:], v.GetBool(config.Music))
		},
	}
	cmd.Flags().IntP("port", "p", 2222, "port the gibson serves SSH on")
	cmd.Flags().Bool("music", true, "play the soundtrack along")
	return cmd
}

// watch runs the intro on this terminal, and the music next to it until
// the intro ends.
func watch(ctx context.Context, ssh, port, host string, words []string, withMusic bool) error {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	defer func() { cancel(); wg.Wait() }()
	if withMusic {
		wg.Go(func() { listen(ctx, ssh, port, host) })
	}

	intro := exec.CommandContext(ctx, ssh, append([]string{"-t", "-p", port, host}, words...)...) //nolint:gosec // the user's own ssh, to the host they named
	intro.Stdin, intro.Stdout, intro.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := intro.Run()
	if exit := new(exec.ExitError); errors.As(err, &exit) {
		return fmt.Errorf("ssh exited with status %d", exit.ExitCode())
	}
	return err
}

// listen plays the host's music stream until ctx is done. It never asks
// for anything, so it can't fight the intro over the terminal: when ssh
// would have to ask (a host it doesn't know yet, a password), the intro's
// ssh does, and listen tries again once that's settled.
func listen(ctx context.Context, ssh, port, host string) {
	for range 4 {
		c := exec.CommandContext(ctx, ssh, "-T", "-o", "BatchMode=yes", "-p", port, host, music.Command) //nolint:gosec // as above
		out, err := c.StdoutPipe()
		if err != nil || c.Start() != nil {
			return
		}
		err = music.Listen(ctx, out)
		_ = c.Wait()
		if ctx.Err() != nil || err == nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
