// +build !windows

package atreugo

import (
	"net"
	"os"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	if s.cfg.Reuseport {
		return reuseport.Listen(s.cfg.Network, s.cfg.Addr) // nolint:wrapcheck
	}

	if s.cfg.Network == "unix" {
		if err := os.Remove(s.cfg.Addr); err != nil && !os.IsNotExist(err) {
			return nil, wrapErrorf(err, "unexpected error when trying to remove unix socket file %q", s.cfg.Addr)
		}
	}

	ln, err := net.Listen(s.cfg.Network, s.cfg.Addr)
	if err != nil {
		return nil, wrapError(err, "failed to announce on the local network address")
	}

	if s.cfg.Network == "unix" {
		if err := s.cfg.chmodUnixSocket(s.cfg.Addr); err != nil {
			return nil, err
		}
	}

	if tcpln, ok := ln.(*net.TCPListener); ok {
		ln = &tcpKeepaliveListener{
			netTCPListener:  &tcpListener{TCPListener: tcpln},
			keepalive:       s.cfg.TCPKeepalive,
			keepalivePeriod: s.cfg.TCPKeepalivePeriod,
		}
	}

	return ln, nil
}
