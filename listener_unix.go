// +build !windows

package atreugo

import (
	"net"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener(addr string) net.Listener {
	ln, err := reuseport.Listen(network, addr)
	if err == nil {
		return ln
	}
	s.log.Warningf("Can not use reuseport listener %s", err)

	s.log.Infof("Trying with net listener")
	ln, err = net.Listen(network, addr)
	panicOnError(err)

	return ln
}
