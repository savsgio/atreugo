package atreugo

import (
	"bufio"
	"bytes"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var testAtreugoConfig = &Config{
	LogLevel: "fatal",
}

var random = func(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func Test_New(t *testing.T) {
	type args struct {
		logLevel string
	}
	type want struct {
		logLevel string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "DefaultLogLevel",
			args: args{
				logLevel: "",
			},
			want: want{
				logLevel: logger.INFO,
			},
		},
		{
			name: "CustomLogLevel",
			args: args{
				logLevel: logger.WARNING,
			},
			want: want{
				logLevel: logger.WARNING,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.args.logLevel}
			if New(cfg); cfg.LogLevel != tt.want.logLevel {
				t.Errorf("Log level = %v, want %v", cfg.LogLevel, tt.want.logLevel)
			}
		})
	}
}

func TestAtreugoServer(t *testing.T) {
	type args struct {
		viewFn        View
		middlewareFns []Middleware
	}
	type want struct {
		statusCode        int
		viewCalled        bool
		middleWareCounter int
	}

	viewCalled := false
	middleWareCounter := 0

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "AllOk",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        fasthttp.StatusOK,
				viewCalled:        true,
				middleWareCounter: 1,
			},
		},
		{
			name: "FirstMiddlewareError",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						return 403, errors.New("Bad request")
					},
					func(ctx *RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        403,
				viewCalled:        false,
				middleWareCounter: 0,
			},
		},
		{
			name: "SecondMiddlewareError",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
					func(ctx *RequestCtx) (int, error) {
						return 403, errors.New("Bad request")
					},
				},
			},
			want: want{
				statusCode:        403,
				viewCalled:        false,
				middleWareCounter: 1,
			},
		},
		{
			name: "ViewError",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					viewCalled = true
					return errors.New("Fake error")
				},
				middlewareFns: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        fasthttp.StatusInternalServerError,
				viewCalled:        true,
				middleWareCounter: 1,
			},
		},
	}

	for _, tt := range tests {
		viewCalled = false
		middleWareCounter = 0

		t.Run(tt.name, func(t *testing.T) {
			s := New(testAtreugoConfig)
			s.UseMiddleware(tt.args.middlewareFns...)
			s.Path("GET", "/", tt.args.viewFn)

			ln := fasthttputil.NewInmemoryListener()

			serveCh := make(chan error, 1)
			go func() {
				serveCh <- s.Serve(ln)
			}()

			clientCh := make(chan error)
			go func() {
				c, err := ln.Dial()
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if _, err = c.Write([]byte("GET / HTTP/1.1\r\nHost: TestServer\r\n\r\n")); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				br := bufio.NewReader(c)
				var resp fasthttp.Response
				if err = resp.Read(br); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if resp.StatusCode() != tt.want.statusCode {
					t.Fatalf("Unexpected status code: '%d', want '%d'", resp.StatusCode(), tt.want.statusCode)
				}

				if viewCalled != tt.want.viewCalled {
					t.Errorf("View called = %v, want %v", viewCalled, tt.want.viewCalled)
				}

				if middleWareCounter != tt.want.middleWareCounter {
					t.Errorf("Middleware call counter = %v, want %v", middleWareCounter, tt.want.middleWareCounter)
				}

				clientCh <- c.Close()
			}()

			select {
			case err := <-serveCh:
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			case err := <-clientCh:
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}
		})
	}
}

func TestAtreugo_Serve(t *testing.T) {
	cfg := &Config{LogLevel: "fatal"}
	s := New(cfg)

	host := "InmemoryListener"
	port := 0
	lnAddr := "InmemoryListener"

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
		if cfg.Host != host {
			t.Errorf("Config.Host = %s, want %s", cfg.Host, host)
		}

		if cfg.Port != port {
			t.Errorf("Config.Port = %d, want %d", cfg.Port, port)
		}

		if s.lnAddr != lnAddr {
			t.Errorf("Atreugo.lnAddr = %s, want %s", s.lnAddr, lnAddr)
		}
	}
}

func TestAtreugo_ServeGracefully(t *testing.T) {
	cfg := &Config{LogLevel: "fatal"}
	s := New(cfg)

	host := "InmemoryListener"
	port := 0
	lnAddr := "InmemoryListener"

	ln := fasthttputil.NewInmemoryListener()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ServeGracefully(ln)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("Unexpected error: %v", err)
	case <-time.After(100 * time.Millisecond):
		if !cfg.GracefulShutdown {
			t.Errorf("Config.GracefulShutdown = %v, want %v", cfg.GracefulShutdown, true)
		}

		if cfg.Host != host {
			t.Errorf("Config.Host = %s, want %s", cfg.Host, host)
		}

		if cfg.Port != port {
			t.Errorf("Config.Port = %d, want %d", cfg.Port, port)
		}

		if s.lnAddr != lnAddr {
			t.Errorf("Atreugo.lnAddr = %s, want %s", s.lnAddr, lnAddr)
		}
	}
}

func TestAtreugo_Static(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		getPanic bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ok",
			args: args{
				path: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(testAtreugoConfig)
			s.Static(tt.args.path)

			if s.router.NotFound == nil {
				t.Error("Static files not configure")
			}
		})
	}
}

func TestAtreugo_Path(t *testing.T) {
	type args struct {
		method string
		url    string
		viewFn View
	}
	type want struct {
		getPanic bool
	}
	testViewFn := func(ctx *RequestCtx) error {
		return nil
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "GET",
			args: args{
				method: "GET",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "HEAD",
			args: args{
				method: "HEAD",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "OPTIONS",
			args: args{
				method: "OPTIONS",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "POST",
			args: args{
				method: "POST",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "PUT",
			args: args{
				method: "PUT",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "PATCH",
			args: args{
				method: "PATCH",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "DELETE",
			args: args{
				method: "DELETE",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "InvalidMethod",
			args: args{
				method: "FAKE",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()

				if tt.want.getPanic && r == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.getPanic && r != nil {
					t.Errorf("Unexpected panic")
				}
			}()

			s := New(testAtreugoConfig)
			s.Path(tt.args.method, tt.args.url, tt.args.viewFn)
		})
	}
}

func TestAtreugo_UseMiddleware(t *testing.T) {
	middlewareFns := []Middleware{
		func(ctx *RequestCtx) (int, error) {
			return 403, errors.New("Bad request")
		},
		func(ctx *RequestCtx) (int, error) {
			return 0, nil
		},
	}

	s := New(testAtreugoConfig)
	s.UseMiddleware(middlewareFns...)

	if len(s.middlewares) != len(middlewareFns) {
		t.Errorf("Middlewares are not registered")
	}

}

func TestAtreugo_SetLogOutput(t *testing.T) {
	s := New(&Config{LogLevel: "info"})
	output := new(bytes.Buffer)

	s.SetLogOutput(output)
	s.log.Info("Test")

	if len(output.Bytes()) <= 0 {
		t.Error("SetLogOutput() log output was not changed")
	}
}

func TestAtreugo_ListenAndServe(t *testing.T) {
	type args struct {
		host      string
		port      int
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
				host:      "localhost",
				port:      random(8000, 9000),
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
				host:      "localhost",
				port:      random(8000, 9000),
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
				host:      "localhost",
				port:      random(8000, 9000),
				tlsEnable: true,
			},
			want: want{
				getErr: true,
			},
		},
		{
			name: "InvalidAddr",
			args: args{
				host: "0101",
				port: 999999999999999999,
			},
			want: want{
				getErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(&Config{
				Host:             tt.args.host,
				Port:             tt.args.port,
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
				s.server.Shutdown()
				if tt.want.getErr {
					t.Error("Error expected")
				}
			}
		})
	}
}

// Benchmarks
func Benchmark_handler(b *testing.B) {
	s := New(testAtreugoConfig)
	viewFn := func(ctx *RequestCtx) error {
		return nil
	}
	ctx := new(fasthttp.RequestCtx)
	h := s.handler(viewFn)

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		h(ctx)
	}
}
