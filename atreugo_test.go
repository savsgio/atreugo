package atreugo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"
	"unicode"

	"github.com/atreugo/mock"
	"github.com/savsgio/gotils/bytes"
	"github.com/savsgio/gotils/strings"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/valyala/fasthttp/prefork"
)

var notConfigFasthttpFields = []string{
	"Handler",
	"ErrorHandler",
}

func Test_New(t *testing.T) { //nolint:funlen,gocognit,gocyclo
	type args struct {
		network                 string
		gracefulShutdown        bool
		gracefulShutdownSignals []os.Signal
		jsonMarshalFunc         JSONMarshalFunc
		notFoundView            View
		methodNotAllowedView    View
		panicView               PanicView
	}

	type want struct {
		gracefulShutdownSignals []os.Signal
		jsonMarshalFunc         JSONMarshalFunc
		notFoundView            bool
		methodNotAllowedView    bool
		panicView               bool
		err                     bool
	}

	jsonMarshalFunc := func(_ io.Writer, _ any) error {
		return nil
	}
	notFoundView := func(_ *RequestCtx) error {
		return nil
	}
	methodNotAllowedView := func(_ *RequestCtx) error {
		return nil
	}

	panicErr := errors.New("error")
	panicView := func(ctx *RequestCtx, err any) {
		ctx.Error(fmt.Sprint(err), fasthttp.StatusInternalServerError)
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
				gracefulShutdownSignals: nil,
				jsonMarshalFunc:         defaultJSONMarshalFunc,
				panicView:               false,
			},
		},
		{
			name: "DefaultGracefulShutdown",
			args: args{
				gracefulShutdown: true,
			},
			want: want{
				gracefulShutdownSignals: defaultGracefulShutdownSignals,
				jsonMarshalFunc:         defaultJSONMarshalFunc,
				notFoundView:            false,
				methodNotAllowedView:    false,
				panicView:               false,
			},
		},
		{
			name: "Custom",
			args: args{
				network:                 "unix",
				gracefulShutdown:        true,
				gracefulShutdownSignals: []os.Signal{syscall.SIGKILL},
				jsonMarshalFunc:         jsonMarshalFunc,
				notFoundView:            notFoundView,
				methodNotAllowedView:    methodNotAllowedView,
				panicView:               panicView,
			},
			want: want{
				gracefulShutdownSignals: []os.Signal{syscall.SIGKILL},
				jsonMarshalFunc:         jsonMarshalFunc,
				notFoundView:            true,
				methodNotAllowedView:    true,
				panicView:               true,
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
			t.Helper()

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
				Network:                 tt.args.network,
				GracefulShutdown:        tt.args.gracefulShutdown,
				GracefulShutdownSignals: tt.args.gracefulShutdownSignals,
				JSONMarshalFunc:         tt.args.jsonMarshalFunc,
				NotFoundView:            tt.args.notFoundView,
				MethodNotAllowedView:    tt.args.methodNotAllowedView,
				PanicView:               tt.args.panicView,
			}
			s := New(cfg)

			if !isEqual(s.cfg.chmodUnixSocketFunc, chmodFileToSocket) {
				t.Errorf("Config.chmodUnixSocketFunc func = %p, want %p", s.cfg.chmodUnixSocketFunc, chmodFileToSocket)
			}

			if !isEqual(s.cfg.newPreforkServerFunc, newPreforkServer) {
				t.Errorf("Config.newPreforkServerFunc func = %p, want %p", s.cfg.newPreforkServerFunc, newPreforkServer)
			}

			if !isEqual(s.cfg.Logger, defaultLogger) {
				t.Errorf("Logger == %p, want %p", s.cfg.Logger, defaultLogger)
			}

			if !reflect.DeepEqual(tt.want.gracefulShutdownSignals, s.cfg.GracefulShutdownSignals) {
				t.Errorf(
					"GracefulShutdownSignals = %v, want %v",
					s.cfg.GracefulShutdownSignals, tt.want.gracefulShutdownSignals,
				)
			}

			if !isEqual(s.cfg.JSONMarshalFunc, tt.want.jsonMarshalFunc) {
				t.Errorf("JSONMarshalFunc == %p, want %p", s.cfg.JSONMarshalFunc, tt.want.jsonMarshalFunc)
			}

			if tt.want.notFoundView != (s.router.NotFound != nil) {
				t.Error("NotFound handler is not setted")
			}

			if tt.want.methodNotAllowedView != (s.router.MethodNotAllowed != nil) {
				t.Error("MethodNotAllowed handler is not setted")
			}

			if !isEqual(s.cfg.ErrorView, defaultErrorView) {
				t.Errorf("Error view == %p, want %p", s.cfg.ErrorView, defaultErrorView)
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

			if s.router == nil {
				t.Fatal("Atreugo router instance is nil")
			}

			if s.router.GlobalOPTIONS != nil {
				t.Error("GlobalOPTIONS handler is not nil")
			}
		})
	}
}

func Test_newFasthttpServer(t *testing.T) { //nolint:funlen
	cfg := Config{
		Name: "test",
		HeaderReceived: func(_ *fasthttp.RequestHeader) fasthttp.RequestConfig {
			return fasthttp.RequestConfig{}
		},
		ContinueHandler:                    func(_ *fasthttp.RequestHeader) bool { return true },
		Concurrency:                        rand.Int(),                              // nolint:gosec
		ReadBufferSize:                     rand.Int(),                              // nolint:gosec
		WriteBufferSize:                    rand.Int(),                              // nolint:gosec
		ReadTimeout:                        time.Duration(rand.Int()),               // nolint:gosec
		WriteTimeout:                       time.Duration(rand.Int()),               // nolint:gosec
		IdleTimeout:                        time.Duration(rand.Int()),               // nolint:gosec
		MaxConnsPerIP:                      rand.Int(),                              // nolint:gosec
		MaxRequestsPerConn:                 rand.Int(),                              // nolint:gosec
		MaxKeepaliveDuration:               time.Duration(rand.Int()) * time.Second, // nolint:gosec
		MaxIdleWorkerDuration:              time.Duration(rand.Int()) * time.Second, // nolint:gosec
		TCPKeepalivePeriod:                 time.Duration(rand.Int()),               // nolint:gosec
		MaxRequestBodySize:                 rand.Int(),                              // nolint:gosec
		DisableKeepalive:                   true,
		TCPKeepalive:                       true,
		ReduceMemoryUsage:                  true,
		GetOnly:                            true,
		DisablePreParseMultipartForm:       true,
		LogAllErrors:                       true,
		SecureErrorLogMessage:              true,
		DisableHeaderNamesNormalizing:      true,
		SleepWhenConcurrencyLimitsExceeded: time.Duration(rand.Int()), // nolint:gosec
		NoDefaultServerHeader:              true,
		NoDefaultDate:                      true,
		NoDefaultContentType:               true,
		KeepHijackedConns:                  true,
		CloseOnShutdown:                    true,
		StreamRequestBody:                  true,
		ConnState:                          func(net.Conn, fasthttp.ConnState) {},
		Logger:                             testLog,
		TLSConfig:                          &tls.Config{ServerName: "test", MinVersion: tls.VersionTLS13},
		FormValueFunc:                      func(_ *fasthttp.RequestCtx, _ string) []byte { return nil },
	}

	srv := newFasthttpServer(cfg)
	if isNil(srv) {
		t.Fatal("newFasthttpServer() == nil")
	}

	fasthttpServerType := reflect.TypeOf(fasthttp.Server{})
	configType := reflect.TypeOf(Config{})

	fasthttpServerValue := reflect.ValueOf(*srv) // nolint:govet
	configValue := reflect.ValueOf(cfg)

	for i := 0; i < fasthttpServerType.NumField(); i++ {
		field := fasthttpServerType.Field(i)

		if !unicode.IsUpper(rune(field.Name[0])) { // Check if the field is public
			continue
		} else if strings.Include(notConfigFasthttpFields, field.Name) {
			continue
		}

		_, exist := configType.FieldByName(field.Name)
		if !exist {
			t.Fatalf("The field '%s' does not exist in atreugo.Config", field.Name)
		}

		v1 := fmt.Sprint(fasthttpServerValue.FieldByName(field.Name).Interface())
		v2 := fmt.Sprint(configValue.FieldByName(field.Name).Interface())

		if v1 != v2 {
			t.Errorf("fasthttp.Server.%s == %s, want %s", field.Name, v1, v2)
		}
	}

	if !isNil(srv.Handler) {
		t.Error("fasthttp.Server.Handler must be nil")
	}

	if !isEqual(srv.Logger, testLog) {
		t.Errorf("fasthttp.Server.Logger == %p, want %p", srv.Logger, testLog)
	}
}

func TestAtreugo_handler(t *testing.T) { // nolint:funlen,gocognit
	type args struct {
		cfg   Config
		hosts []string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Default",
			args: args{
				cfg: Config{},
			},
		},
		{
			name: "Compress",
			args: args{
				cfg: Config{Compress: true},
			},
		},
		{
			name: "MultiHost",
			args: args{
				cfg:   Config{},
				hosts: []string{"localhost", "example.com"},
			},
		},
		{
			name: "MultiHostCompress",
			args: args{
				cfg:   Config{Compress: true},
				hosts: []string{"localhost", "example.com"},
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			testView := func(ctx *RequestCtx) error {
				return ctx.JSONResponse(JSON{"data": bytes.Rand(make([]byte, 300))})
			}
			testPath := "/"

			s := New(tt.args.cfg)
			s.GET(testPath, testView)

			for _, hostname := range tt.args.hosts {
				vHost := s.NewVirtualHost(hostname)
				vHost.GET(testPath, testView)
			}

			handler := s.handler()

			if handler == nil {
				t.Errorf("handler is nil")
			}

			newHostname := string(bytes.Rand(make([]byte, 10))) + ".com"

			hosts := tt.args.hosts
			hosts = append(hosts, newHostname)

			for _, hostname := range hosts {
				for _, path := range []string{testPath, "/notfound"} {
					ctx := new(fasthttp.RequestCtx)
					ctx.Request.Header.Set(fasthttp.HeaderAcceptEncoding, "gzip")
					ctx.Request.Header.Set(fasthttp.HeaderHost, hostname)
					ctx.Request.URI().SetHost(hostname)
					ctx.Request.SetRequestURI(path)

					handler(ctx)

					statusCode := ctx.Response.StatusCode()
					wantStatusCode := fasthttp.StatusOK

					if path != testPath {
						wantStatusCode = fasthttp.StatusNotFound
					}

					if statusCode != wantStatusCode {
						t.Errorf("Host %s - Path %s, Status code == %d, want %d", hostname, path, statusCode, wantStatusCode)
					}

					if wantStatusCode == fasthttp.StatusNotFound {
						continue
					}

					if tt.args.cfg.Compress && len(ctx.Response.Header.Peek(fasthttp.HeaderContentEncoding)) == 0 {
						t.Errorf("The header '%s' is not setted", fasthttp.HeaderContentEncoding)
					}
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
			t.Helper()

			s := New(testConfig)
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
		Logger:      testLog,
		ReadTimeout: 1 * time.Second,
	}
	s := New(cfg)

	c := &mock.Conn{ErrRead: errors.New("Read error")}
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.ServeConn(c)
	}()

	time.Sleep(100 * time.Millisecond)

	if err := s.engine.Shutdown(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err == nil {
		t.Fatalf("Expected error: %v", err)
	}

	if s.engine.Handler == nil {
		t.Error("Atreugo.engine.Handler is nil")
	}
}

func TestAtreugo_Serve(t *testing.T) {
	s := New(testConfig)

	ln := fasthttputil.NewInmemoryListener()
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Serve(ln)
	}()

	time.Sleep(500 * time.Millisecond)

	if err := s.engine.Shutdown(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lnAddr := ln.Addr().String(); s.cfg.Addr != lnAddr {
		t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
	}

	lnNetwork := ln.Addr().Network()
	if s.cfg.Network != lnNetwork {
		t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
	}

	if s.engine.Handler == nil {
		t.Error("Atreugo.engine.Handler is nil")
	}
}

func TestAtreugo_NewVirtualHost(t *testing.T) { //nolint:funlen
	hostname := "localhost"

	s := New(testConfig)

	if s.virtualHosts != nil {
		t.Error("Atreugo.virtualHosts must be nil before register a new virtual host")
	}

	vHost := s.NewVirtualHost(hostname)
	if isNil(vHost) {
		t.Fatal("Atreugo.NewVirtualHost() returned a nil router")
	}

	if !isEqual(vHost.router.NotFound, s.router.NotFound) {
		t.Errorf("VirtualHost router.NotFound == %p, want %p", vHost.router.NotFound, s.router.NotFound)
	}

	if !isEqual(vHost.router.MethodNotAllowed, s.router.MethodNotAllowed) {
		t.Errorf(
			"VirtualHost router.MethodNotAllowed == %p, want %p",
			vHost.router.MethodNotAllowed,
			s.router.MethodNotAllowed,
		)
	}

	if !isEqual(vHost.router.PanicHandler, s.router.PanicHandler) {
		t.Errorf("VirtualHost router.PanicHandler == %p, want %p", vHost.router.PanicHandler, s.router.PanicHandler)
	}

	if h := s.virtualHosts[hostname]; isNil(h) {
		t.Error("The new virtual host is not registeded")
	}

	type conflictArgs struct {
		hostnames  []string
		wantErrMsg string
	}

	conflictHosts := []conflictArgs{
		{
			hostnames:  []string{hostname},
			wantErrMsg: "a router is already registered for virtual host: " + hostname,
		},
		{
			hostnames:  []string{},
			wantErrMsg: "at least 1 hostname is required",
		},
		{
			hostnames:  []string{"localhost", "localhost"},
			wantErrMsg: "a router is already registered for virtual host: " + hostname,
		},
	}

	for _, test := range conflictHosts {
		tt := test

		err := catchPanic(func() {
			s.NewVirtualHost(tt.hostnames...)
		})

		if err == nil {
			t.Error("Expected panic when a virtual host is duplicated")
		}

		if err != tt.wantErrMsg {
			t.Errorf("Error string == %s, want %s", err, tt.wantErrMsg)
		}
	}
}

func TestAtreugo_Shutdown(t *testing.T) {
	s := New(testConfig)

	ln := fasthttputil.NewInmemoryListener()
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Serve(ln)
	}()

	time.Sleep(500 * time.Millisecond)

	if err := s.Shutdown(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lnAddr := ln.Addr().String(); s.cfg.Addr != lnAddr {
		t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
	}

	lnNetwork := ln.Addr().Network()
	if s.cfg.Network != lnNetwork {
		t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
	}

	if s.engine.Handler == nil {
		t.Error("Atreugo.engine.Handler is nil")
	}
}

func TestAtreugo_ShutdownWithContext(t *testing.T) {
	s := New(testConfig)

	ln := fasthttputil.NewInmemoryListener()
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.Serve(ln)
	}()

	time.Sleep(500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := s.ShutdownWithContext(ctx); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lnAddr := ln.Addr().String(); s.cfg.Addr != lnAddr {
		t.Errorf("Atreugo.Config.Addr = %s, want %s", s.cfg.Addr, lnAddr)
	}

	lnNetwork := ln.Addr().Network()
	if s.cfg.Network != lnNetwork {
		t.Errorf("Atreugo.Config.Network = %s, want %s", s.cfg.Network, lnNetwork)
	}

	if s.engine.Handler == nil {
		t.Error("Atreugo.engine.Handler is nil")
	}
}

// Benchmarks.
func Benchmark_Handler(b *testing.B) {
	s := New(testConfig)
	s.GET("/plaintext", func(_ *RequestCtx) error { return nil })
	s.GET("/json", func(_ *RequestCtx) error { return nil })
	s.GET("/db", func(_ *RequestCtx) error { return nil })
	s.GET("/queries", func(_ *RequestCtx) error { return nil })
	s.GET("/cached-worlds", func(_ *RequestCtx) error { return nil })
	s.GET("/fortunes", func(_ *RequestCtx) error { return nil })
	s.GET("/updates", func(_ *RequestCtx) error { return nil })

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/updates")

	handler := s.handler()

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			handler(ctx)
		}
	})
}
