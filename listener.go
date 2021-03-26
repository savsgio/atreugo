package atreugo

import "net"

func (ln *tcpListener) AcceptTCP() (netTCPConn, error) {
	return ln.TCPListener.AcceptTCP() // nolint:wrapcheck
}

func (ln *tcpKeepaliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if err := tc.SetKeepAlive(ln.keepalive); err != nil {
		tc.Close() //nolint:errcheck

		return nil, err //nolint:wrapcheck
	}

	if ln.keepalivePeriod > 0 {
		if err := tc.SetKeepAlivePeriod(ln.keepalivePeriod); err != nil {
			tc.Close() //nolint:errcheck

			return nil, err //nolint:wrapcheck
		}
	}

	return tc, nil
}
