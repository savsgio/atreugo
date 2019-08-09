// +build windows

package atreugo

import (
	"net"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	return net.Listen(s.cfg.Network, s.lnAddr)
}
