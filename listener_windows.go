//go:build windows
// +build windows

package atreugo

import (
	"net"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	if s.cfg.Reuseport {
		return reuseport.Listen(s.cfg.Network, s.cfg.Addr)
	}

	ln, err := net.Listen(s.cfg.Network, s.cfg.Addr)
	if err != nil {
		return nil, wrapError(err, "failed to announce on the local network address")
	}

	return ln, nil
}
