// +build windows

package atreugo

import (
	"net"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	return net.Listen(network, s.lnAddr)
}
