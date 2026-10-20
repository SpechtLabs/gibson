// Package command builds the gibson command tree: the root command, which
// plays the intro in the local terminal, and serve, render, music, denied
// and healthcheck. Every subcommand lives in its own sub-package with a
// NewCommand constructor. The root owns the persistent look flags and loads the config
// file before any command runs; each command binds its own flags to their
// config keys when it runs.
package command

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	"github.com/spechtlabs/gibson/cmd/gibson/command/denied"
	"github.com/spechtlabs/gibson/cmd/gibson/command/healthcheck"
	"github.com/spechtlabs/gibson/cmd/gibson/command/music"
	"github.com/spechtlabs/gibson/cmd/gibson/command/render"
	"github.com/spechtlabs/gibson/cmd/gibson/command/serve"
	"github.com/spechtlabs/gibson/internal/app"
	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/scene"
)

// lookFlags maps the root's persistent flags to their config keys.
var lookFlags = map[string]string{
	"theme":    config.Theme,
	"specimen": config.Specimen,
	"vessel":   config.Vessel,
	"tagline":  config.Tagline,
	"fps":      config.FPS,
}

// NewCommand returns the gibson root command with every subcommand
// attached. version is logged when serve starts.
func NewCommand(version string) *cobra.Command {
	v := config.New()
	cmd := &cobra.Command{
		Use:   "gibson",
		Short: "Play the neon SpechtLabs intro",
		Long: `gibson plays a retro neon "SpechtLabs" intro in the terminal, or serves it
over SSH with gibson serve.

Every flag can also be set in a config file (--config, or gibson.yaml in .,
$XDG_CONFIG_HOME/gibson or /etc/gibson) or a GIBSON_ environment variable,
e.g. GIBSON_THEME or GIBSON_SSH_LISTEN. Flags win over the environment,
which wins over the file.`,
		Example: `# Play the intro
gibson --theme specht --vessel dodeca

# Serve it over SSH
gibson serve --listen :2222`,
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("config")
			if err := config.Load(v, path); err != nil {
				return err
			}
			return config.Bind(v, cmd.Flags(), lookFlags)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Bind(v, cmd.Flags(), map[string]string{"start": config.Start, "pausable": config.Pausable}); err != nil {
				return err
			}
			look, err := config.Look(v)
			if err != nil {
				return err
			}
			sc, err := look.New()
			if err != nil {
				return err
			}
			fps := v.GetInt(config.FPS)
			m := app.New(sc, fps, v.GetFloat64(config.Start), v.GetBool(config.Pausable))
			_, err = tea.NewProgram(m, tea.WithFPS(fps)).Run()
			return err
		},
	}

	pf := cmd.PersistentFlags()
	pf.String("config", "", "config file (default: gibson.yaml in ., $XDG_CONFIG_HOME/gibson, /etc/gibson)")
	pf.String("theme", "synth", "color theme: "+strings.Join(scene.Themes(), " | "))
	pf.String("specimen", "bird", "what floats in the vessel: "+strings.Join(scene.Specimens(), " | "))
	pf.String("vessel", "hex", "shape of the containment vessel: "+strings.Join(scene.Vessels(), " | "))
	pf.String("tagline", scene.DefaultTagline, "text typed under the title")
	pf.Int("fps", 30, "frames per second")

	cmd.Flags().Float64("start", 0, "start the sequence at this time (seconds)")
	cmd.Flags().Bool("pausable", false, "space toggles pause instead of quitting (TEMP)")

	cmd.AddCommand(
		serve.NewCommand(v, version),
		render.NewCommand(v),
		denied.NewCommand(v),
		healthcheck.NewCommand(v),
		music.NewCommand(),
	)
	return cmd
}

// Execute runs the gibson command line on os.Args with fang's styling for
// help and errors, and returns the process's exit status: 1 when the
// command failed, after printing why, and 0 otherwise.
func Execute(ctx context.Context, version string) int {
	if err := fang.Execute(ctx, NewCommand(version), fang.WithVersion(version)); err != nil {
		return 1
	}
	return 0
}
