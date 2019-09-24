// +build !windows

package atreugo

import (
	"fmt"
	"net"
	"os"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	if s.cfg.Reuseport {
		return reuseport.Listen(s.cfg.Network, s.cfg.Addr)
	}

	if s.cfg.Network == "unix" {
		if err := os.Remove(s.cfg.Addr); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("unexpected error when trying to remove unix socket file %q: %s", s.cfg.Addr, err)
		}
	}

	ln, err := net.Listen(s.cfg.Network, s.cfg.Addr)

	if s.cfg.Network == "unix" {
		if err := os.Chmod(s.cfg.Addr, s.cfg.socketFileMode); err != nil {
			return nil, fmt.Errorf("cannot chmod %#o for %q: %s", s.cfg.socketFileMode, s.cfg.Addr, err)
		}
	}

	return ln, err
}
