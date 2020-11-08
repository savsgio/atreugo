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
			return nil, fmt.Errorf("unexpected error when trying to remove unix socket file %q: %w", s.cfg.Addr, err)
		}
	}

	ln, err := net.Listen(s.cfg.Network, s.cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to announce on the local network address: %w", err)
	}

	if s.cfg.Network == "unix" {
		if err := os.Chmod(s.cfg.Addr, os.ModeSocket); err != nil {
			return nil, fmt.Errorf("cannot chmod %#o for %q: %w", os.ModeSocket, s.cfg.Addr, err)
		}
	}

	return ln, nil
}
