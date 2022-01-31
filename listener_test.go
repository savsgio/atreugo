package atreugo

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/atreugo/mock"
)

type mockTCPListener struct {
	mock.Listener

	errAcceptTCP          error
	errSetKeepAlive       error
	errSetKeepAlivePeriod error
}

type mockTCPConn struct {
	mock.Conn

	errSetKeepAlive       error
	errSetKeepAlivePeriod error
}

func (conn *mockTCPConn) SetKeepAlive(keepalive bool) error {
	return conn.errSetKeepAlive
}

func (conn *mockTCPConn) SetKeepAlivePeriod(d time.Duration) error {
	return conn.errSetKeepAlivePeriod
}

func (ln *mockTCPListener) AcceptTCP() (netTCPConn, error) {
	if ln.errAcceptTCP != nil {
		return nil, ln.errAcceptTCP
	}

	conn := &mockTCPConn{
		errSetKeepAlive:       ln.errSetKeepAlive,
		errSetKeepAlivePeriod: ln.errSetKeepAlivePeriod,
	}

	return conn, nil
}

func TestTCPListener_AcceptTCP(t *testing.T) {
	tcpln := new(net.TCPListener)
	ln := &tcpListener{TCPListener: tcpln}

	conn, _ := ln.AcceptTCP() // nolint:ifshort

	if _, ok := conn.(*net.TCPConn); !ok {
		t.Errorf("conn type == %T, want %T", conn, &net.TCPConn{})
	}
}

func TestTCPKeepaliveListener_Accept(t *testing.T) { // nolint:funlen
	type args struct {
		errAcceptTCP          error
		errSetKeepAlive       error
		errSetKeepAlivePeriod error
	}

	type want struct {
		err error
	}

	errAcceptTCP := errors.New("error")
	errSetKeepAlive := errors.New("error")
	errSetKeepAlivePeriod := errors.New("errSetKeepAlivePeriod")
	keepalivePeriod := 10 * time.Second

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ok",
			args: args{
				errSetKeepAlive:       nil,
				errSetKeepAlivePeriod: nil,
			},
			want: want{err: nil},
		},
		{
			name: "AcceptTCPError",
			args: args{
				errAcceptTCP:          errAcceptTCP,
				errSetKeepAlive:       nil,
				errSetKeepAlivePeriod: nil,
			},
			want: want{err: errAcceptTCP},
		},
		{
			name: "SetKeepAliveError",
			args: args{
				errAcceptTCP:          nil,
				errSetKeepAlive:       errSetKeepAlive,
				errSetKeepAlivePeriod: nil,
			},
			want: want{err: errSetKeepAlive},
		},
		{
			name: "SetKeepAlivePeriodError",
			args: args{
				errAcceptTCP:          nil,
				errSetKeepAlive:       nil,
				errSetKeepAlivePeriod: errSetKeepAlivePeriod,
			},
			want: want{err: errSetKeepAlivePeriod},
		},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			t.Helper()

			tcpln := &mockTCPListener{
				errAcceptTCP:          test.args.errAcceptTCP,
				errSetKeepAlive:       test.args.errSetKeepAlive,
				errSetKeepAlivePeriod: test.args.errSetKeepAlivePeriod,
			}

			ln := tcpKeepaliveListener{
				netTCPListener:  tcpln,
				keepalivePeriod: keepalivePeriod,
			}

			conn, err := ln.Accept()
			if !errors.Is(err, test.want.err) {
				t.Errorf("Error == %s, want %s", err, test.want.err)
			}

			if test.want.err == nil {
				_, ok := conn.(*mockTCPConn)
				if !ok {
					t.Errorf("conn type == %T, want %T", conn, &mockTCPConn{})
				}
			}
		})
	}
}
