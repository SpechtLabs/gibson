package sshserver

import (
	"runtime/debug"
	"strings"

	"charm.land/log/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
)

// (In a file of its own: sshserver.go imports wish's recover package,
// which hides the builtin there.)

// commandRouter hands sessions without a pty that ask for one of cmds to
// it, guarded against panics as the program is, and everything else on.
func commandRouter(cmds map[string]ssh.Handler) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			args := s.Command()
			if _, _, pty := s.Pty(); pty || len(args) == 0 || cmds[args[0]] == nil {
				next(s)
				return
			}
			defer func() {
				if r := recover(); r != nil {
					log.Error("command panicked", "command", args[0], "panic", r, "stack", string(debug.Stack()))
				}
			}()
			cmds[args[0]](s)
		}
	}
}

// needTerm turns away sessions without a pty, which the program needs,
// saying how to get one: ssh allocates one for a command only with -t.
func needTerm() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if _, _, ok := s.Pty(); ok {
				next(s)
				return
			}
			wish.Fatalf(s, "gibson needs a terminal; ssh allocates one for a command only with -t:\n  ssh -t -p <port> <host> %s\n", strings.Join(s.Command(), " "))
		}
	}
}
