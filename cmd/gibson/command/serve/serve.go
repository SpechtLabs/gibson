// Package serve implements gibson serve, which serves the intro over SSH.
package serve

import (
	"time"

	"charm.land/log/v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/server"
)

// flags maps serve's flags to their config keys.
var flags = map[string]string{
	"listen":       config.Listen,
	"host-key":     config.HostKey,
	"max-conns":    config.MaxConns,
	"max-sessions": config.MaxSessions,
	"rate":         config.Rate,
	"burst":        config.Burst,
	"max-duration": config.MaxDuration,
	"pausable":     config.Pausable,
}

// NewCommand returns the serve command, which reads its settings from v
// and logs version on start.
func NewCommand(v *viper.Viper, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the intro over SSH",
		Long: `Serves the intro as an SSH app: every connection gets its own scene. Words
passed as the SSH command pick the look for that session, overriding the
configured one, e.g. ssh -t host -p 2222 dodeca specht.

The server is open to anyone but bounded everywhere: open connections,
concurrent sessions, new sessions per IP, connection lifetime and terminal
size.`,
		Example: `gibson serve --listen :2222 --host-key /data/ssh_host_ed25519`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Bind(v, cmd.Flags(), flags); err != nil {
				return err
			}
			if f := v.ConfigFileUsed(); f != "" {
				log.Info("loaded config", "file", f)
			}
			look, err := config.Look(v)
			if err != nil {
				return err
			}
			return server.Serve(server.Options{
				Addr:     v.GetString(config.Listen),
				HostKey:  v.GetString(config.HostKey),
				Defaults: look,
				FPS:      v.GetInt(config.FPS),
				Pausable: v.GetBool(config.Pausable),
				Limits: server.Limits{
					MaxConns:    v.GetInt(config.MaxConns),
					MaxSessions: v.GetInt(config.MaxSessions),
					PerIPRate:   v.GetFloat64(config.Rate),
					PerIPBurst:  v.GetInt(config.Burst),
					MaxDuration: v.GetDuration(config.MaxDuration),
				},
				Version: version,
			})
		},
	}

	f := cmd.Flags()
	f.String("listen", ":2222", "address to serve SSH on")
	f.String("host-key", ".ssh/id_ed25519", "path of the SSH host key (generated if missing)")
	f.Int("max-conns", 128, "maximum open connections")
	f.Int("max-sessions", 32, "maximum concurrent sessions")
	f.Float64("rate", 0.5, "new sessions per second allowed per IP")
	f.Int("burst", 5, "burst of new sessions allowed per IP")
	f.Duration("max-duration", 15*time.Minute, "maximum lifetime of a connection")
	f.Bool("pausable", false, "space toggles pause instead of quitting (TEMP)")
	return cmd
}
