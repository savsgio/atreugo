// +build windows

package atreugo

import (
	"net"
)

func (s *Atreugo) getListener(addr string) net.Listener {
	ln, err := net.Listen(network, addr)
	panicOnError(err)

	return ln
}
