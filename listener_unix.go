// +build !windows

package atreugo

import (
	"net"
	"runtime"

	"github.com/valyala/fasthttp/reuseport"
)

func (s *Atreugo) getListener() (net.Listener, error) {
	if runtime.NumCPU() > 1 {
		ln, err := reuseport.Listen(network, s.lnAddr)
		if err == nil {
			return ln, nil
		}
		s.log.Warning("Can not use reuseport, using default Listener")
	}

	return net.Listen(network, s.lnAddr)
}
