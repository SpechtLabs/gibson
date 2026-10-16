// Package server serves the intro as a Wish SSH app.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	bm "charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"charm.land/wish/v2/ratelimiter"
	"charm.land/wish/v2/recover"
	"github.com/charmbracelet/colorprofile"
	"golang.org/x/time/rate"

	"github.com/spechtlabs/gibson/internal/app"
	"github.com/spechtlabs/gibson/internal/scene"
	"github.com/spechtlabs/gibson/internal/theme"
)

// Limits bound the server. Every viewer costs a full render per frame, so the server is open to
// anyone but bounded everywhere: connections, sessions, session rate per
// IP, session length and terminal size (see [app.MaxCols]).
type Limits struct {
	MaxConns    int           // open TCP connections, handshakes included
	MaxSessions int           // concurrent sessions (one connection can open many)
	PerIPRate   float64       // new sessions per second per IP
	PerIPBurst  int           // burst on top of PerIPRate
	MaxDuration time.Duration // absolute connection lifetime
}

// Options configure [Serve].
type Options struct {
	Addr     string
	HostKey  string        // path, generated if missing
	Defaults scene.Options // the look of a session that names none
	FPS      int
	Pausable bool // TEMP: space toggles pause
	Limits   Limits
	Version  string // logged on start
}

// Serve runs the intro as an SSH app: every connection gets its own scene
// and Bubble Tea program. Words passed as the SSH command pick the look for
// that session, e.g. `ssh -t host -p 2222 dodeca specht`.
func Serve(o Options) error {
	lim := o.Limits
	var sessions atomic.Int64
	srv, err := wish.NewServer(
		wish.WithAddress(o.Addr),
		wish.WithHostKeyPath(o.HostKey),
		wish.WithMaxTimeout(lim.MaxDuration),
		func(s *ssh.Server) error {
			s.HandshakeTimeout = 10 * time.Second
			s.ConnCallback = connLimiter(lim.MaxConns)
			return nil
		},
		wish.WithMiddleware(
			// composed first to last, so the last one here runs first
			recover.Middleware(
				bm.MiddlewareWithProgramHandler(func(sess ssh.Session) *tea.Program {
					opt := sessionOptions(o.Defaults, sess.Command())
					m := app.New(scene.New(opt), o.FPS, 0, o.Pausable)
					opts := append(bm.MakeOptions(sess),
						tea.WithFPS(o.FPS),
						// ssh clients rarely forward COLORTERM, and the whole look
						// depends on 24-bit gradients, so don't let it downsample
						tea.WithColorProfile(colorprofile.TrueColor),
					)
					return tea.NewProgram(m, opts...)
				}),
			),
			activeterm.Middleware(), // bubbletea needs a pty
			sessionLimiter(&sessions, lim.MaxSessions),
			ratelimiter.Middleware(ratelimiter.NewRateLimiter(rate.Limit(lim.PerIPRate), lim.PerIPBurst, 4096)),
			logging.Middleware(),
		),
	)
	if err != nil {
		return err
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Info("starting SSH server", "version", o.Version, "addr", o.Addr, "connect", "ssh -p "+port(o.Addr)+" localhost")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// connLimiter drops new TCP connections once max are open.
func connLimiter(maxConns int) ssh.ConnCallback {
	var open atomic.Int64
	return func(_ ssh.Context, c net.Conn) net.Conn {
		if open.Add(1) > int64(maxConns) {
			open.Add(-1)
			log.Warn("connection limit reached, dropping", "remote", c.RemoteAddr())
			return nil
		}
		return &countedConn{Conn: c, release: func() { open.Add(-1) }}
	}
}

type countedConn struct {
	net.Conn
	once    sync.Once
	release func()
}

func (c *countedConn) Close() error {
	c.once.Do(c.release)
	return c.Conn.Close()
}

// sessionLimiter turns sessions away once max are running.
func sessionLimiter(n *atomic.Int64, maxSessions int) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			defer n.Add(-1)
			if n.Add(1) > int64(maxSessions) {
				wish.Fatalln(s, "the gibson is at capacity, try again in a bit")
				return
			}
			next(s)
		}
	}
}

// sessionOptions overrides the server defaults with any theme / vessel /
// specimen names given as the SSH command.
func sessionOptions(o scene.Options, args []string) scene.Options {
	for _, a := range args {
		switch {
		case slices.Contains(theme.Names(), a):
			o.Theme, _ = theme.Get(a)
		case slices.Contains(scene.Vessels, a):
			o.Vessel = a
		case slices.Contains(scene.Specimens, a):
			o.Specimen = a
		}
	}
	return o
}

func port(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil {
		return p
	}
	return addr
}
