package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Healthcheck succeeds when something on addr speaks SSH. The image has no
// shell, so the container HEALTHCHECK runs the binary itself.
func Healthcheck(addr string) error {
	host, p, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "" {
		host = "127.0.0.1"
	}
	c, err := net.DialTimeout("tcp", net.JoinHostPort(host, p), 3*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(3 * time.Second))
	banner, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(banner, "SSH-2.0-") {
		return fmt.Errorf("unexpected banner %q", banner)
	}
	return nil
}
