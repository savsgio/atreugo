// +build !windows

package atreugo

import (
	"bytes"
	"errors"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/atreugo/mock"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/valyala/fasthttp/prefork"
)

func Test_IsPreforkChild(t *testing.T) {
	if IsPreforkChild() != prefork.IsChild() {
		t.Errorf("IsPreforkChild() == %v, want %v", IsPreforkChild(), prefork.IsChild())
	}
}

func TestAtreugo_newPreforkServer(t *testing.T) {
	cfg := Config{
		LogLevel:         "fatal",
		GracefulShutdown: false,
	}

	s := New(cfg)
	sPrefork := s.newPreforkServer()

	if sPrefork.Network != s.cfg.Network {
		t.Errorf("Prefork.Network == %s, want %s", sPrefork.Network, s.cfg.Network)
	}

	if sPrefork.Reuseport != s.cfg.Reuseport {
		t.Errorf("Prefork.Reuseport == %v, want %v", sPrefork.Reuseport, s.cfg.Reuseport)
	}

	recoverThreshold := runtime.GOMAXPROCS(0) / 2
	if sPrefork.RecoverThreshold != recoverThreshold {
		t.Errorf("Prefork.RecoverThreshold == %d, want %d", sPrefork.RecoverThreshold, recoverThreshold)
	}

	if !isEqual(sPrefork.Logger, s.log) {
		t.Errorf("Prefork.Logger == %p, want %p", sPrefork.Logger, s.log)
	}

	if !isEqual(sPrefork.ServeFunc, s.Serve) {
		t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.Serve)
	}

	// With graceful shutdown
	cfg.GracefulShutdown = true

	s = New(cfg)
	sPrefork = s.newPreforkServer()

	if isEqual(sPrefork.ServeFunc, s.Serve) {
		t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.ServeGracefully)
	}

	if !isEqual(sPrefork.ServeFunc, s.ServeGracefully) {
		t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.ServeGracefully)
	}
}

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
			ln := &mock.Listener{
				Listener:    fasthttputil.NewInmemoryListener(),
				AcceptError: tt.args.lnAcceptError,
				CloseError:  tt.args.lnCloseError,
			}
			defer ln.Listener.Close()

			logOutput := &bytes.Buffer{}

			cfg := Config{LogLevel: "fatal"}
			s := New(cfg)
			s.SetLogOutput(logOutput)

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

			if s.server.ReadTimeout != defaultReadTimeout {
				t.Errorf("fasthttp.Server.ReadTimeout = %v, want %v", s.server.ReadTimeout, defaultReadTimeout)
			}

			if s.cfg.ReadTimeout != defaultReadTimeout {
				t.Errorf("Config.ReadTimeout = %v, want %v", s.cfg.ReadTimeout, defaultReadTimeout)
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
	type args struct {
		addr      string
		graceful  bool
		tlsEnable bool
		prefork   bool
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
				graceful:  false,
				tlsEnable: false,
				prefork:   false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "GracefulOk",
			args: args{
				addr:      "localhost:8081",
				graceful:  true,
				tlsEnable: false,
				prefork:   false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "PreforkError",
			args: args{
				addr:      "invalid",
				graceful:  false,
				tlsEnable: false,
				prefork:   true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "PreforkGracefulError",
			args: args{
				addr:      "invalid",
				graceful:  true,
				tlsEnable: false,
				prefork:   true,
			},
			want: want{
				getErr: true,
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
				Addr:             tt.args.addr,
				LogLevel:         "error",
				TLSEnable:        tt.args.tlsEnable,
				GracefulShutdown: tt.args.graceful,
				Prefork:          tt.args.prefork,
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
