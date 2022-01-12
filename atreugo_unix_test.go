//go:build !windows
// +build !windows

package atreugo

import (
	"bytes"
	"errors"
	"log"
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
		Logger:           testLog,
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

	if !isEqual(sPrefork.Logger, s.cfg.Logger) {
		t.Errorf("Prefork.Logger == %p, want %p", sPrefork.Logger, s.cfg.Logger)
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
				Addr:             "localhost:8081",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "GracefulOk",
			args: Config{
				Addr:             "localhost:8081",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          false,
			},
			want: want{
				getErr: false,
			},
		},
		{
			name: "PreforkError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: false,
				TLSEnable:        false,
				Prefork:          true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "PreforkGracefulError",
			args: Config{
				Addr:             "invalid",
				GracefulShutdown: true,
				TLSEnable:        false,
				Prefork:          true,
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
			t.Helper()

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
