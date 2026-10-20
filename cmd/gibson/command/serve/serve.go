// Package serve implements gibson serve, which serves the intro over SSH.
package serve

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/ssh"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/app"
	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/denied"
	"github.com/spechtlabs/gibson/internal/music"
	"github.com/spechtlabs/gibson/pkg/sshserver"
)

// flags maps serve's flags to their config keys.
var flags = map[string]string{
	"listen":         config.Listen,
	"host-key":       config.HostKey,
	"max-conns":      config.MaxConns,
	"max-sessions":   config.MaxSessions,
	"rate":           config.Rate,
	"burst":          config.Burst,
	"max-duration":   config.MaxDuration,
	"proxy-protocol": config.ProxyProto,
	"pausable":       config.Pausable,
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
ssh -p 2222 host music | ffplay - streams the 8-bit soundtrack as a WAV.

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
			fps, pausable := v.GetInt(config.FPS), v.GetBool(config.Pausable)
			srv, err := sshserver.New(
				sshserver.WithAddress(v.GetString(config.Listen)),
				sshserver.WithHostKeyPath(v.GetString(config.HostKey)),
				sshserver.WithProxyProtocol(v.GetBool(config.ProxyProto)),
				sshserver.WithLimits(sshserver.Limits{
					MaxConns:    v.GetInt(config.MaxConns),
					MaxSessions: v.GetInt(config.MaxSessions),
					PerIPRate:   v.GetFloat64(config.Rate),
					PerIPBurst:  v.GetInt(config.Burst),
					MaxDuration: v.GetDuration(config.MaxDuration),
					Handshake:   10 * time.Second,
				}),
				sshserver.WithModel(func(sess ssh.Session) tea.Model {
					// words passed as the SSH command pick the look; they're
					// all known names, so this can't fail
					sc, _ := look.Override(sess.Command()).New()
					return app.New(sc, fps, 0, pausable)
				}),
				sshserver.WithProgramOptions(
					tea.WithFPS(fps),
					// ssh clients rarely forward COLORTERM, and the whole look
					// depends on 24-bit gradients, so don't let it downsample
					tea.WithColorProfile(colorprofile.TrueColor),
				),
				// a turned-away viewer gets the scene frozen and a line from
				// Hackers; gibson denied previews it
				sshserver.WithCapacityHandler(denied.Handler(look, app.MaxCols, app.MaxRows)),
				// SSH carries no sound: ssh host music | ffplay -
				sshserver.WithCommand(music.Command, music.Handler),
			)
			if err != nil {
				return err
			}
			log.Info("gibson", "version", version)
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return srv.Run(ctx)
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
	f.Bool("proxy-protocol", false, "read the client's address from a PROXY protocol header, behind a load balancer that sends one (only when nothing else can reach the server)")
	f.Bool("pausable", false, "space toggles pause instead of quitting (TEMP)")
	return cmd
}
