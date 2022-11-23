//go:build windows
// +build windows

package atreugo

import (
	"testing"
)

func TestAtreugo_getListener(t *testing.T) { // nolint:funlen
	type want struct {
		addr    string
		network string
		err     bool
	}

	tests := []struct {
		name string
		args Config
		want want
	}{
		{
			name: "Ok",
			args: Config{
				Addr: "127.0.0.1:8000",
			},
			want: want{
				addr:    "127.0.0.1:8000",
				network: "tcp",
				err:     false,
			},
		},
		{
			name: "Reuseport",
			args: Config{
				Addr:      "127.0.0.1:8000",
				Network:   "tcp4",
				Reuseport: true,
			},
			want: want{
				addr:    "127.0.0.1:8000",
				network: "tcp",
				err:     false,
			},
		},
		{
			name: "Unix",
			args: Config{
				Addr:    "/tmp/test.sock",
				Network: "unix",
			},
			want: want{
				err: true,
			},
		},
		{
			name: "Error",
			args: Config{
				Network: "fake",
			},
			want: want{
				err: true,
			},
		},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()

				if tt.want.err && err == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.err && err != nil {
					t.Errorf("Unexpected panic: %v", err)
				}
			}()

			s := New(tt.args)

			ln, err := s.getListener()
			if err != nil {
				panic(err)
			}

			defer ln.Close()

			lnAddress := ln.Addr().String()
			if lnAddress != tt.want.addr {
				t.Errorf("Listener address: '%s', want '%s'", lnAddress, tt.want.addr)
			}

			lnNetwork := ln.Addr().Network()
			if lnNetwork != tt.want.network {
				t.Errorf("Listener network: '%s', want '%s'", lnNetwork, tt.want.network)
			}
		})
	}
}
