//go:build windows
// +build windows

package atreugo

import (
	"errors"
	"testing"
	"time"
)

func TestAtreugo_ListenAndServe(t *testing.T) { //nolint:funlen
	type want struct {
		err error
	}

	tests := []struct {
		name string
		args Config
		want want
	}{
		{
			name: "NormalOk",
			args: Config{
				Addr:      "localhost:8081",
				TLSEnable: false,
				Prefork:   false,
				Reuseport: false,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "ReuseportOk",
			args: Config{
				Addr:      "localhost:8081",
				TLSEnable: false,
				Prefork:   false,
				Reuseport: true,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "NormalError",
			args: Config{
				Addr:      "invalid",
				TLSEnable: false,
				Prefork:   false,
				Reuseport: false,
			},
			want: want{
				err: errors.New(
					"failed to announce on the local network address: listen tcp4: address invalid: missing port in address",
				),
			},
		},
		{
			name: "PreforkError",
			args: Config{
				Addr:      "invalid",
				TLSEnable: false,
				Prefork:   true,
				Reuseport: false,
			},
			want: want{
				err: errors.New("prefork error"),
			},
		},
		{
			name: "ReuseportError",
			args: Config{
				Addr:      "invalid",
				TLSEnable: false,
				Prefork:   false,
				Reuseport: true,
			},
			want: want{
				err: errors.New("listen tcp4: address invalid: missing port in address"),
			},
		},
		{
			name: "TLSError",
			args: Config{
				Addr:      "localhost:8081",
				TLSEnable: true,
			},
			want: want{
				err: errors.New("cert or key has not provided"),
			},
		},
	}

	for _, test := range tests {
		tt := test
		tt.args.Logger = testLog

		waitTime := 200 * time.Millisecond

		s := New(tt.args)
		s.cfg.newPreforkServerFunc = func(s *Atreugo) preforkServer {
			return newPreforkServerMock(s, tt.want.err)
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			errCh := make(chan error, 1)
			go func() {
				errCh <- s.ListenAndServe()
			}()

			var err error

			select {
			case err = <-errCh:
			case <-time.After(waitTime):
				if err := s.engine.Shutdown(); err != nil {
					t.Errorf("Error shutting down the server: %+v", err)
				}
			}

			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}

			wantErrMessage := ""
			if tt.want.err != nil {
				wantErrMessage = tt.want.err.Error()
			}

			if errMessage != wantErrMessage {
				t.Errorf("Unexpected error: %s, want: %s", errMessage, wantErrMessage)
			}
		})
	}
}
