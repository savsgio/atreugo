// +build !windows

package atreugo

import (
	"net"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	if s.cfg.Reuseport {
		return reuseport.Listen(s.cfg.Network, s.cfg.Addr)
	}

	return net.Listen(s.cfg.Network, s.cfg.Addr)
}
