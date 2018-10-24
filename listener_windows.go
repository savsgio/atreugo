// +build windows

package atreugo

import (
	"net"
)

func (s *Atreugo) getListener() (net.Listener, err) {
	return net.Listen(network, s.lnAddr)
}
