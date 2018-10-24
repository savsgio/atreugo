package atreugo

import "testing"

func TestAtreugo_getListener(t *testing.T) {
	type args struct {
		addr string
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
			name: "Error",
			args: args{
				addr: "fake",
			},
			want: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(testAtreugoConfig)
			s.lnAddr = tt.args.addr

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
