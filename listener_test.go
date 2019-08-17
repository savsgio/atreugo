package atreugo

import (
	"fmt"
	"testing"
)

func TestAtreugo_getListener(t *testing.T) {
	type args struct {
		host      string
		port      int
		network   string
		reuseport bool
	}
	type want struct {
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
				host: "127.0.0.1",
				port: 8000,
			},
			want: want{
				network: "tcp",
				err:     false,
			},
		},
		{
			name: "Reuseport",
			args: args{
				host:      "127.0.0.1",
				port:      8000,
				network:   "tcp4",
				reuseport: true,
			},
			want: want{
				network: "tcp",
				err:     false,
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Host:      tt.args.host,
				Port:      tt.args.port,
				LogLevel:  "fatal",
				Reuseport: tt.args.reuseport,
			}
			if tt.args.network != "" {
				cfg.Network = tt.args.network
			}

			s := New(cfg)

			s.lnAddr = fmt.Sprintf("%s:%d", tt.args.host, tt.args.port)

			defer func() {
				r := recover()

				if tt.want.err && r == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.err && r != nil {
					t.Errorf("Unexpected panic")
				}
			}()

			ln, err := s.getListener()
			if err != nil {
				panic(err)
			}

			lnAddress := ln.Addr().String()
			if lnAddress != s.lnAddr {
				t.Errorf("Listener address: '%s', want '%s'", lnAddress, s.lnAddr)
			}

			lnNetwork := ln.Addr().Network()
			if lnNetwork != tt.want.network {
				t.Errorf("Listener network: '%s', want '%s'", lnNetwork, tt.want.network)
			}

			ln.Close()
		})
	}
}
