package atreugo

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"
	"unicode"

	"github.com/atreugo/mock"
	logger "github.com/savsgio/go-logger/v2"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
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

func TestAtreugo_SetLogOutput(t *testing.T) {
	s := New(Config{LogLevel: "info"})
	output := new(bytes.Buffer)

	s.SetLogOutput(output)
	s.log.Info("Test")

	if len(output.Bytes()) == 0 {
		t.Error("SetLogOutput() log output was not changed")
	}
}
