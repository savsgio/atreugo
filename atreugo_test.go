package atreugo

import (
	"bytes"
	"errors"
	"net"
	"reflect"
	"syscall"
	"testing"
	"time"
	"unicode"

	logger "github.com/savsgio/go-logger"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var testAtreugoConfig = &Config{
	LogLevel: "fatal",
}

var notConfigFasthttpFields = []string{"Handler", "ErrorHandler", "TCPKeepalive", "TCPKeepalivePeriod", "Logger"}

type mockListener struct {
	ln net.Listener

	acceptError error
	closeError  error

	acceptCalled bool
	closeCalled  bool
	addrCalled   bool
}

func (m *mockListener) Accept() (net.Conn, error) {
	m.acceptCalled = true

	if m.acceptError != nil {
		return nil, m.acceptError
	}

	return m.ln.Accept()
}

func (m *mockListener) Close() error {
	m.closeCalled = true

	if m.closeError != nil {
		return m.closeError
	}

	return m.ln.Close()
}

func (m *mockListener) Addr() net.Addr {
	m.addrCalled = true
	return m.ln.Addr()
}

func newMockListener(ln net.Listener, acceptError, closeError error) *mockListener {
	return &mockListener{ln: ln, acceptError: acceptError, closeError: closeError}
}

func Test_New(t *testing.T) { //nolint:funlen,gocognit
	type args struct {
		network              string
		logLevel             string
		globalOPTIONSView    View
		notFoundView         View
		methodNotAllowedView View
		panicView            PanicView
	}

	type want struct {
		logLevel             string
		globalOPTIONSView    bool
		notFoundView         bool
		methodNotAllowedView bool
		panicView            bool
		err                  bool
	}

	globalOPTIONSView := func(ctx *RequestCtx) error {
		return nil
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
				globalOPTIONSView:    false,
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
				globalOPTIONSView:    globalOPTIONSView,
				notFoundView:         notFoundView,
				methodNotAllowedView: methodNotAllowedView,
				panicView:            panicView,
			},
			want: want{
				logLevel:             logger.WARNING,
				globalOPTIONSView:    true,
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

			cfg := &Config{
				Network:              tt.args.network,
				LogLevel:             tt.args.logLevel,
				GlobalOPTIONS:        tt.args.globalOPTIONSView,
				NotFoundView:         tt.args.notFoundView,
				MethodNotAllowedView: tt.args.methodNotAllowedView,
				PanicView:            tt.args.panicView,
			}
			s := New(cfg)

			if cfg.LogLevel != tt.want.logLevel {
				t.Errorf("Log level = %v, want %v", cfg.LogLevel, tt.want.logLevel)
			}

			if s.router == nil {
				t.Fatal("Atreugo router instance is nil")
			}

			if tt.want.globalOPTIONSView != (s.router.GlobalOPTIONS != nil) {
				t.Error("GlobalOPTIONS handler is not setted")
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

func Test_fasthttpServer(t *testing.T) { //nolint:funlen
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
			cfg := &Config{
				LogLevel: "fatal",
				Compress: tt.args.compress,
			}
			srv := fasthttpServer(cfg, handler, testLog)

			if (reflect.ValueOf(handler).Pointer() == reflect.ValueOf(srv.Handler).Pointer()) == tt.want.compress {
				t.Error("The handler has not been wrapped by compression handler")
			}
		})
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

			if s.router.HandleOPTIONS != tt.want.v {
				t.Errorf("Router.HandleOPTIONS == %v, want %v", s.router.HandleOPTIONS, tt.want.v)
			}
		})
	}
}

func TestAtreugo_Serve(t *testing.T) {
	cfg := &Config{LogLevel: "fatal"}
	s := New(cfg)

	ln := fasthttputil.NewInmemoryListener()
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Serve(ln)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		lnAddr := ln.Addr().String()
		if s.cfg.Addr != lnAddr {
			t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
		}

		lnNetwork := ln.Addr().Network()
		if s.cfg.Network != lnNetwork {
			t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
		}
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
			ln := newMockListener(
				fasthttputil.NewInmemoryListener(), tt.args.lnAcceptError, tt.args.lnCloseError,
			)
			defer ln.ln.Close()

			logOutput := &bytes.Buffer{}

			cfg := &Config{LogLevel: "fatal"}
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

			if !cfg.GracefulShutdown {
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
	s := New(&Config{LogLevel: "info"})
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
			s := New(&Config{
				Addr:             tt.args.addr,
				LogLevel:         "error",
				TLSEnable:        tt.args.tlsEnable,
				GracefulShutdown: tt.args.graceful,
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
