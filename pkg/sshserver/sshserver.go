// Package sshserver serves a Bubble Tea program over SSH with Wish: every
// session gets a model of its own. It's open to anyone, so it's bounded
// everywhere instead: open connections, concurrent sessions, new sessions
// per IP, handshake time and connection lifetime.
package sshserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
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
	"golang.org/x/time/rate"
)

// Limits bound the server.
type Limits struct {
	MaxConns    int           // open TCP connections, handshakes included
	MaxSessions int           // concurrent sessions (one connection can open many)
	PerIPRate   float64       // new sessions per second per IP
	PerIPBurst  int           // burst on top of PerIPRate
	MaxDuration time.Duration // absolute connection lifetime
	Handshake   time.Duration // time allowed for the SSH handshake
}

// DefaultLimits are the limits a [Server] has unless [WithLimits] says
// otherwise.
func DefaultLimits() Limits {
	return Limits{
		MaxConns:    128,
		MaxSessions: 32,
		PerIPRate:   0.5,
		PerIPBurst:  5,
		MaxDuration: 15 * time.Minute,
		Handshake:   10 * time.Second,
	}
}

// Server is an SSH server for a Bubble Tea program.
type Server struct {
	addr     string
	hostKey  string
	limits   Limits
	model    func(ssh.Session) tea.Model
	progOpts []tea.ProgramOption
	full     ssh.Handler
	proxied  bool
	commands map[string]ssh.Handler

	srv *ssh.Server
}

// Option configures a [Server].
type Option func(*Server)

// WithAddress sets the address to listen on (default ":2222").
func WithAddress(addr string) Option { return func(s *Server) { s.addr = addr } }

// WithHostKeyPath sets where the host key lives; it's generated there if
// it's missing (default ".ssh/id_ed25519").
func WithHostKeyPath(path string) Option { return func(s *Server) { s.hostKey = path } }

// WithLimits sets the server's limits (default [DefaultLimits]).
func WithLimits(l Limits) Option { return func(s *Server) { s.limits = l } }

// WithModel sets the model each session runs, made from the session (its
// command, its user, its pty). Required.
func WithModel(f func(ssh.Session) tea.Model) Option { return func(s *Server) { s.model = f } }

// WithProgramOptions adds options to every session's program, after the
// ones that wire it to the session.
func WithProgramOptions(opts ...tea.ProgramOption) Option {
	return func(s *Server) { s.progOpts = append(s.progOpts, opts...) }
}

// WithProxyProtocol reads the PROXY protocol header (v1 or v2) a proxy in
// front of the server puts at the start of each connection, so limits and
// logs see the client's address rather than the proxy's. A connection
// without a header is still served, from its own address, so only enable
// it when nothing but the proxy can reach the server: anything else could
// claim any address.
func WithProxyProtocol(on bool) Option { return func(s *Server) { s.proxied = on } }

// WithCapacityMessage sets what a session turned away at MaxSessions reads.
func WithCapacityMessage(msg string) Option {
	return WithCapacityHandler(func(sess ssh.Session) { wish.Fatalln(sess, msg) })
}

// WithCapacityHandler sets what a session turned away at MaxSessions gets
// instead of a message. It runs before the session has been checked for a
// pty, and the session ends when it returns. It doesn't count against
// MaxSessions, so keep it short and cheap.
func WithCapacityHandler(h ssh.Handler) Option { return func(s *Server) { s.full = h } }

// WithCommand runs h instead of the program for sessions whose command
// starts with name (ssh host name), with or without a pty. It's for what
// isn't a terminal UI, such as a stream to pipe into something else. These
// sessions count against MaxSessions like any other.
func WithCommand(name string, h ssh.Handler) Option {
	return func(s *Server) {
		if s.commands == nil {
			s.commands = map[string]ssh.Handler{}
		}
		s.commands[name] = h
	}
}

// ErrNoModel is returned by [New] without [WithModel].
var ErrNoModel = errors.New("sshserver: no model; use WithModel")

// New returns a server.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		addr:    ":2222",
		hostKey: ".ssh/id_ed25519",
		limits:  DefaultLimits(),
	}
	WithCapacityMessage("the server is at capacity, try again in a bit")(s)
	for _, o := range opts {
		o(s)
	}
	if s.model == nil {
		return nil, ErrNoModel
	}
	lim := s.limits
	var sessions atomic.Int64
	srv, err := wish.NewServer(
		wish.WithAddress(s.addr),
		wish.WithHostKeyPath(s.hostKey),
		wish.WithMaxTimeout(lim.MaxDuration),
		func(srv *ssh.Server) error {
			srv.HandshakeTimeout = lim.Handshake
			srv.ConnCallback = connLimiter(lim.MaxConns)
			srv.EnableProxyProtocol = s.proxied
			return nil
		},
		wish.WithMiddleware(
			// composed first to last, so the last one here runs first
			recover.Middleware(
				bm.MiddlewareWithProgramHandler(func(sess ssh.Session) *tea.Program {
					opts := append(bm.MakeOptions(sess), s.progOpts...)
					return tea.NewProgram(s.model(sess), opts...)
				}),
			),
			activeterm.Middleware(), // bubbletea needs a pty
			commandRouter(s.commands),
			sessionLimiter(&sessions, lim.MaxSessions, s.full),
			ratelimiter.Middleware(ratelimiter.NewRateLimiter(rate.Limit(lim.PerIPRate), lim.PerIPBurst, 4096)),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("sshserver: %w", err)
	}
	s.srv = srv
	return s, nil
}

// Run serves until ctx is done, then shuts down like [Server.Shutdown]
// with ten seconds for open sessions to end.
func (s *Server) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	errc := make(chan error, 1)
	var wg sync.WaitGroup
	logger.Info("starting SSH server", "addr", s.addr, "connect", "ssh -p "+port(s.addr)+" localhost")
	// ends when ListenAndServe returns, which Shutdown below makes it do
	wg.Go(func() { errc <- s.srv.ListenAndServe() })
	defer wg.Wait()

	select {
	case err := <-errc:
		if errors.Is(err, ssh.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	logger.Info("stopping SSH server")
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.Shutdown(sctx)
}

// Shutdown stops accepting connections and waits for open sessions to end
// until ctx is done.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
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
func sessionLimiter(n *atomic.Int64, maxSessions int, full ssh.Handler) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			defer n.Add(-1)
			if n.Add(1) > int64(maxSessions) {
				full(s)
				return
			}
			next(s)
		}
	}
}

func port(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil {
		return p
	}
	return addr
}
