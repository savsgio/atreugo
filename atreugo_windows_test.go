//go:build windows
// +build windows

package atreugo

import (
	"testing"
	"time"
)

func TestAtreugo_newPreforkServer(t *testing.T) {
	cfg := Config{
		Logger:           testLog,
		GracefulShutdown: false,
	}

	s := New(cfg)
	sPrefork := s.newPreforkServer()

	testPerforkServer(t, s, sPrefork)

	if !isEqual(sPrefork.ServeFunc, s.Serve) {
		t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.Serve)
	}
}

func TestAtreugo_ListenAndServe(t *testing.T) { //nolint:funlen
	type want struct {
		getErr bool
	}

	tests := []struct {
		name string
		args Config
		want want
	}{
		{
			name: "NormalOk",
			args: Config{
				Addr:      "localhost:8083",
				TLSEnable: false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "Reuseport",
			args: Config{
				Addr:      "localhost:8083",
				TLSEnable: false,
				Reuseport: true,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "ReuseportError",
			args: Config{
				Addr:      "invalid",
				TLSEnable: false,
				Reuseport: true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "TLSError",
			args: Config{
				Addr:      "localhost:8081",
				TLSEnable: true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "InvalidAddr",
			args: Config{
				Addr: "0101:999999999999999999",
			},
			want: want{
				getErr: true,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			tt.args.Logger = testLog

			s := New(tt.args)

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
				if err := s.engine.Shutdown(); err != nil {
					t.Errorf("Error shutting down the server %+v", err)
				}
				if tt.want.getErr {
					t.Error("Error expected")
				}
			}
		})
	}
}
