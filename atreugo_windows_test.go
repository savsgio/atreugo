// +build windows

package atreugo

import (
	"testing"
	"time"
)

func TestAtreugo_ListenAndServe(t *testing.T) { //nolint:funlen
	type args struct {
		addr      string
		tlsEnable bool
		// reuseport bool
	}

	type want struct {
		getErr bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "NormalOk",
			args: args{
				addr:      "localhost:8081",
				tlsEnable: false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "TLSError",
			args: args{
				addr:      "localhost:8081",
				tlsEnable: true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "InvalidAddr",
			args: args{
				addr: "0101:999999999999999999",
			},
			want: want{
				getErr: true,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			s := New(Config{
				Addr:      tt.args.addr,
				LogLevel:  "error",
				TLSEnable: tt.args.tlsEnable,
			})

			errCh := make(chan error, 1)
			go func() {
				errCh <- s.ListenAndServe()
			}()

			select {
			case err := <-errCh:
				if !tt.want.getErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			case <-time.After(200 * time.Millisecond):
				if err := s.server.Shutdown(); err != nil {
					t.Errorf("Error shutting down the server %+v", err)
				}
				if tt.want.getErr {
					t.Error("Error expected")
				}
			}
		})
	}
}
