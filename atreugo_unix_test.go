//go:build !windows
// +build !windows

package atreugo

import (
	"bytes"
	"errors"
	"log"
	"syscall"
	"testing"
	"time"

	"github.com/atreugo/mock"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestAtreugo_ServeGracefully(t *testing.T) { // nolint:funlen
	type args struct {
		lnAcceptError error
		lnCloseError  error
	}

	type want struct {
		err bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ok",
			args: args{
				lnAcceptError: nil,
				lnCloseError:  nil,
			},
			want: want{
				err: false,
			},
		},
		{
			name: "ServeError",
			args: args{
				lnAcceptError: errors.New("listener accept error"),
				lnCloseError:  nil,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "ShutdownError",
			args: args{
				lnAcceptError: nil,
				lnCloseError:  errors.New("listener close error"),
			},
			want: want{
				err: true,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ln := &mock.Listener{
				Listener:    fasthttputil.NewInmemoryListener(),
				AcceptError: tt.args.lnAcceptError,
				CloseError:  tt.args.lnCloseError,
			}
			defer ln.Listener.Close()

			logOutput := &bytes.Buffer{}
			log := log.New(logOutput, "", log.LstdFlags)

			cfg := Config{Logger: log}
			s := New(cfg)

			errCh := make(chan error, 1)

			go func() {
				errCh <- s.ServeGracefully(ln)
			}()

			time.Sleep(100 * time.Millisecond)

			if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if err := <-errCh; (err == nil) == tt.want.err {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !s.cfg.GracefulShutdown {
				t.Errorf("Config.GracefulShutdown = %v, want %v", cfg.GracefulShutdown, true)
			}

			lnAddr := ln.Addr().String()
			if s.cfg.Addr != lnAddr {
				t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
			}

			lnNetwork := ln.Addr().Network()
			if s.cfg.Network != lnNetwork {
				t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
			}
		})
	}
}

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
				Addr:             "localhost:8081",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        false,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "GracefulOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        false,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "PreforkOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        false,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "ReuseportOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        true,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "GracefulPreforkOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        false,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "GracefulPreforkReuseportOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        true,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "NormalError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        false,
			},
			want: want{
				err: errors.New(
					"failed to announce on the local network address: listen tcp4: address invalid: missing port in address",
				),
			},
		},
		{
			name: "GracefulError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        false,
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
				Addr:             "invalid",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        false,
			},
			want: want{
				err: errors.New("prefork error"),
			},
		},
		{
			name: "ReuseportError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          false,
				Reuseport:        true,
			},
			want: want{
				err: errors.New("address invalid: missing port in address"),
			},
		},
		{
			name: "GracefulPreforkError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        false,
			},
			want: want{
				err: errors.New("graceful prefork error"),
			},
		},
		{
			name: "GracefulPreforkReuseportError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          true,
				Reuseport:        true,
			},
			want: want{
				err: errors.New("graceful prefork reuseport error"),
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
