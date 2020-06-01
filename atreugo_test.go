package atreugo

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"
	"unicode"

	"github.com/atreugo/mock"
	logger "github.com/savsgio/go-logger/v2"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/valyala/fasthttp/prefork"
)

var testAtreugoConfig = Config{
	LogLevel: "fatal",
}

var notConfigFasthttpFields = []string{
	"Handler",
	"ErrorHandler",
	"TCPKeepalive",
	"TCPKeepalivePeriod",
	"Logger",
	"MaxKeepaliveDuration", // Deprecated: Use IdleTimeout instead.
}

func Test_New(t *testing.T) { //nolint:funlen,gocognit
	type args struct {
		network              string
		logLevel             string
		notFoundView         View
		methodNotAllowedView View
		panicView            PanicView
	}

	type want struct {
		logLevel             string
		notFoundView         bool
		methodNotAllowedView bool
		panicView            bool
		err                  bool
	}

	notFoundView := func(ctx *RequestCtx) error {
		return nil
	}
	methodNotAllowedView := func(ctx *RequestCtx) error {
		return nil
	}

	panicErr := errors.New("error")
	panicView := func(ctx *RequestCtx, err interface{}) {
		ctx.Error(panicErr.Error(), fasthttp.StatusInternalServerError)
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Default",
			args: args{},
			want: want{
				logLevel:             logger.INFO,
				notFoundView:         false,
				methodNotAllowedView: false,
				panicView:            false,
			},
		},
		{
			name: "Custom",
			args: args{
				network:              "unix",
				logLevel:             logger.WARNING,
				notFoundView:         notFoundView,
				methodNotAllowedView: methodNotAllowedView,
				panicView:            panicView,
			},
			want: want{
				logLevel:             logger.WARNING,
				notFoundView:         true,
				methodNotAllowedView: true,
				panicView:            true,
			},
		},
		{
			name: "InvalidNetwork",
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
			defer func() {
				r := recover()

				switch {
				case tt.want.err && r == nil:
					t.Errorf("Panic expected")
				case !tt.want.err && r != nil:
					t.Errorf("Unexpected panic")
				}
			}()

			cfg := Config{
				Network:              tt.args.network,
				LogLevel:             tt.args.logLevel,
				NotFoundView:         tt.args.notFoundView,
				MethodNotAllowedView: tt.args.methodNotAllowedView,
				PanicView:            tt.args.panicView,
			}
			s := New(cfg)

			if s.cfg.LogLevel != tt.want.logLevel {
				t.Errorf("Log level = %v, want %v", cfg.LogLevel, tt.want.logLevel)
			}

			if s.router == nil {
				t.Fatal("Atreugo router instance is nil")
			}

			if s.router.GlobalOPTIONS != nil {
				t.Error("GlobalOPTIONS handler is not nil")
			}

			if tt.want.notFoundView != (s.router.NotFound != nil) {
				t.Error("NotFound handler is not setted")
			}

			if tt.want.methodNotAllowedView != (s.router.MethodNotAllowed != nil) {
				t.Error("MethodNotAllowed handler is not setted")
			}

			if tt.want.panicView != (s.router.PanicHandler != nil) {
				t.Error("PanicHandler handler is not setted")
			}

			if tt.args.panicView != nil {
				ctx := new(fasthttp.RequestCtx)
				s.router.PanicHandler(ctx, panicErr)

				if string(ctx.Response.Body()) != panicErr.Error() {
					t.Errorf("Panic handler response == %s, want %s", ctx.Response.Body(), panicErr.Error())
				}
			}
		})
	}
}

func Test_newFasthttpServer(t *testing.T) { //nolint:funlen
	type args struct {
		compress bool
	}

	type want struct {
		compress bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "NotCompress",
			args: args{
				compress: false,
			},
			want: want{
				compress: false,
			},
		},
		{
			name: "Compress",
			args: args{
				compress: true,
			},
			want: want{
				compress: true,
			},
		},
	}

	handler := func(ctx *fasthttp.RequestCtx) {}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				LogLevel: "fatal",
				Compress: tt.args.compress,
			}
			srv := newFasthttpServer(cfg, handler, testLog)

			if (reflect.ValueOf(handler).Pointer() == reflect.ValueOf(srv.Handler).Pointer()) == tt.want.compress {
				t.Error("The handler has not been wrapped by compression handler")
			}
		})
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

func TestAtreugo_ConfigFasthttpFields(t *testing.T) {
	fasthttpServerType := reflect.TypeOf(fasthttp.Server{})
	configType := reflect.TypeOf(Config{})

	for i := 0; i < fasthttpServerType.NumField(); i++ {
		field := fasthttpServerType.Field(i)

		if !unicode.IsUpper(rune(field.Name[0])) { // Check if the field is public
			continue
		} else if gotils.StringSliceInclude(notConfigFasthttpFields, field.Name) {
			continue
		}

		_, exist := configType.FieldByName(field.Name)
		if !exist {
			t.Errorf("The field '%s' does not exist in atreugo.Config", field.Name)
		}
	}
}

func TestAtreugo_RouterConfiguration(t *testing.T) { //nolint:funlen
	type args struct {
		v bool
	}

	type want struct {
		v bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Enable",
			args: args{
				v: true,
			},
			want: want{
				v: true,
			},
		},
		{
			name: "Disable",
			args: args{
				v: false,
			},
			want: want{
				v: false,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			s := New(testAtreugoConfig)
			s.SaveMatchedRoutePath(tt.args.v)
			s.RedirectTrailingSlash(tt.args.v)
			s.RedirectFixedPath(tt.args.v)
			s.HandleMethodNotAllowed(tt.args.v)
			s.HandleOPTIONS(tt.args.v)

			if s.router.SaveMatchedRoutePath != tt.want.v {
				t.Errorf("Router.SaveMatchedRoutePath == %v, want %v", s.router.SaveMatchedRoutePath, tt.want.v)
			}

			if s.router.RedirectTrailingSlash != tt.want.v {
				t.Errorf("Router.RedirectTrailingSlash == %v, want %v", s.router.RedirectTrailingSlash, tt.want.v)
			}

			if s.router.RedirectFixedPath != tt.want.v {
				t.Errorf("Router.RedirectFixedPath == %v, want %v", s.router.RedirectFixedPath, tt.want.v)
			}

			if s.router.HandleMethodNotAllowed != tt.want.v {
				t.Errorf("Router.HandleMethodNotAllowed == %v, want %v", s.router.HandleMethodNotAllowed, tt.want.v)
			}

			if s.router.HandleOPTIONS != false {
				t.Errorf("Router.router.HandleOPTIONS == %v, want %v", s.router.HandleOPTIONS, false)
			}

			if s.handleOPTIONS != tt.want.v {
				t.Errorf("Router.handleOPTIONS == %v, want %v", s.handleOPTIONS, tt.want.v)
			}
		})
	}
}

func TestAtreugo_ServeConn(t *testing.T) {
	cfg := Config{
		LogLevel:    "fatal",
		ReadTimeout: 1 * time.Second,
	}
	s := New(cfg)

	c := &mock.Conn{ErrRead: errors.New("Read error")}
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.ServeConn(c)
	}()

	time.Sleep(100 * time.Millisecond)

	if err := s.server.Shutdown(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err == nil {
		t.Fatalf("Expected error: %v", err)
	}
}

func TestAtreugo_Serve(t *testing.T) {
	cfg := Config{LogLevel: "fatal"}
	s := New(cfg)

	ln := fasthttputil.NewInmemoryListener()
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Serve(ln)
	}()

	time.Sleep(100 * time.Millisecond)

	if err := s.server.Shutdown(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lnAddr := ln.Addr().String()
	if s.cfg.Addr != lnAddr {
		t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
	}

	lnNetwork := ln.Addr().Network()
	if s.cfg.Network != lnNetwork {
		t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
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

func TestAtreugo_SetLogOutput(t *testing.T) {
	s := New(Config{LogLevel: "info"})
	output := new(bytes.Buffer)

	s.SetLogOutput(output)
	s.log.Info("Test")

	if len(output.Bytes()) == 0 {
		t.Error("SetLogOutput() log output was not changed")
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

func Test_IsPreforkChild(t *testing.T) {
	if IsPreforkChild() != prefork.IsChild() {
		t.Errorf("IsPreforkChild() == %v, want %v", IsPreforkChild(), prefork.IsChild())
	}
}
