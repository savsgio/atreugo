// +build windows

package atreugo

import (
	"testing"
)

func TestAtreugo_getListener(t *testing.T) { // nolint:funlen
	type args struct {
		addr      string
		network   string
		reuseport bool
	}

	type want struct {
		addr    string
		network string
		err     bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ok",
			args: args{
				addr: "127.0.0.1:8000",
			},
			want: want{
				addr:    "127.0.0.1:8000",
				network: "tcp",
				err:     false,
			},
		},
		{
			name: "Reuseport",
			args: args{
				addr:      "127.0.0.1:8000",
				network:   "tcp4",
				reuseport: true,
			},
			want: want{
				addr:    "127.0.0.1:8000",
				network: "tcp",
				err:     false,
			},
		},
		{
			name: "Unix",
			args: args{
				addr:    "/tmp/test.sock",
				network: "unix",
			},
			want: want{
				err: true,
			},
		},
		{
			name: "Error",
			args: args{
				network: "fake",
			},
			want: want{
				err: true,
			},
		},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Logger:    testLog,
				Addr:      tt.args.addr,
				Reuseport: tt.args.reuseport,
			}
			if tt.args.network != "" {
				cfg.Network = tt.args.network
			}

			defer func() {
				err := recover()

				if tt.want.err && err == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.err && err != nil {
					t.Errorf("Unexpected panic: %v", err)
				}
			}()

			s := New(cfg)

			ln, err := s.getListener()
			if err != nil {
				panic(err)
			}

			lnAddress := ln.Addr().String()
			if lnAddress != tt.want.addr {
				t.Errorf("Listener address: '%s', want '%s'", lnAddress, tt.want.addr)
			}

			lnNetwork := ln.Addr().Network()
			if lnNetwork != tt.want.network {
				t.Errorf("Listener network: '%s', want '%s'", lnNetwork, tt.want.network)
			}

			ln.Close()
		})
	}
}
