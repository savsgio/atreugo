package atreugo

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
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
	cfg := Config{
		Name: "test",
		HeaderReceived: func(header *fasthttp.RequestHeader) fasthttp.RequestConfig {
			return fasthttp.RequestConfig{}
		},
		Concurrency:                        rand.Int(), // nolint:gosec
		DisableKeepalive:                   true,
		ReadBufferSize:                     rand.Int(),                // nolint:gosec
		WriteBufferSize:                    rand.Int(),                // nolint:gosec
		ReadTimeout:                        time.Duration(rand.Int()), // nolint:gosec
		WriteTimeout:                       time.Duration(rand.Int()), // nolint:gosec
		IdleTimeout:                        time.Duration(rand.Int()), // nolint:gosec
		MaxConnsPerIP:                      rand.Int(),                // nolint:gosec
		MaxRequestsPerConn:                 rand.Int(),                // nolint:gosec
		MaxRequestBodySize:                 rand.Int(),                // nolint:gosec
		ReduceMemoryUsage:                  true,
		GetOnly:                            true,
		DisablePreParseMultipartForm:       true,
		LogAllErrors:                       true,
		DisableHeaderNamesNormalizing:      true,
		SleepWhenConcurrencyLimitsExceeded: time.Duration(rand.Int()), // nolint:gosec
		NoDefaultServerHeader:              true,
		NoDefaultDate:                      true,
		NoDefaultContentType:               true,
		ConnState:                          func(net.Conn, fasthttp.ConnState) {},
		KeepHijackedConns:                  true,
	}

	srv := newFasthttpServer(cfg, testLog)

	if srv == nil {
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
		} else if gotils.StringSliceInclude(notConfigFasthttpFields, field.Name) {
			continue
		}

		_, exist := configType.FieldByName(field.Name)
		if !exist {
			t.Errorf("The field '%s' does not exist in atreugo.Config", field.Name)
		}

		v1 := fmt.Sprint(fasthttpServerValue.FieldByName(field.Name).Interface())
		v2 := fmt.Sprint(configValue.FieldByName(field.Name).Interface())

		if v1 != v2 {
			t.Errorf("fasthttp.Server.%s == %s, want %s", field.Name, v1, v2)
		}
	}

	if srv.Handler != nil {
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
			testView := func(ctx *RequestCtx) error {
				return ctx.JSONResponse(JSON{"data": gotils.RandBytes(make([]byte, 300))})
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

			newHostname := string(gotils.RandBytes(make([]byte, 10))) + ".com"

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

	if s.server.Handler == nil {
		t.Error("Atreugo.server.Handler is nil")
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

	if s.server.Handler == nil {
		t.Error("Atreugo.server.Handler is nil")
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

func TestAtreugo_NewVirtualHost(t *testing.T) { //nolint:funlen
	hostname := "localhost"

	s := New(testAtreugoConfig)

	if s.virtualHosts != nil {
		t.Error("Atreugo.virtualHosts must be nil before register a new virtual host")
	}

	vHost := s.NewVirtualHost(hostname)
	if vHost == nil {
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

	if h := s.virtualHosts[hostname]; h == nil {
		t.Error("The new virtual host is not registeded")
	}

	type conflictArgs struct {
		hostnames  []string
		wantErrMsg string
	}

	conflictHosts := []conflictArgs{
		{
			hostnames:  []string{hostname},
			wantErrMsg: fmt.Sprintf("a router is already registered for virtual host '%s'", hostname),
		},
		{
			hostnames:  []string{},
			wantErrMsg: "At least 1 hostname is required",
		},
		{
			hostnames:  []string{"localhost", "localhost"},
			wantErrMsg: fmt.Sprintf("a router is already registered for virtual host '%s'", hostname),
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

// Benchmarks.
func Benchmark_Handler(b *testing.B) {
	s := New(testAtreugoConfig)
	s.GET("/", func(ctx *RequestCtx) error { return nil })

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/")

	handler := s.handler()

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		handler(ctx)
	}
}
