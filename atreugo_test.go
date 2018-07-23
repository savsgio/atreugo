package atreugo

import (
	"bufio"
	"errors"
	"testing"
	"time"

	"github.com/savsgio/go-logger"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/erikdubbelboer/fasthttp/fasthttputil"
)

var testAtreugoConfig = &Config{
	LogLevel: "error",
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
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
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
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						return 403, errors.New("Bad request")
					},
					func(ctx *fasthttp.RequestCtx) (int, error) {
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
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
					func(ctx *fasthttp.RequestCtx) (int, error) {
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
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return errors.New("Fake error")
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
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

			serverCh := make(chan error, 1)
			go func() {
				err := s.serve(ln)
				serverCh <- err
			}()

			clientCh := make(chan struct{})
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

				if err = c.Close(); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				close(clientCh)
			}()

			select {
			case <-clientCh:
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}

			if err := ln.Close(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			select {
			case <-serverCh:
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}
		})
	}
}

func TestAtreugo_getListener(t *testing.T) {
	type args struct {
		addr string
	}
	type want struct {
		addr     string
		network  string
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
				addr: "127.0.0.1:8000",
			},
			want: want{
				addr:     "127.0.0.1:8000",
				network:  "tcp",
				getPanic: false,
			},
		},
		{
			name: "Error",
			args: args{
				addr: "fake",
			},
			want: want{
				getPanic: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(testAtreugoConfig)

			defer func() {
				r := recover()

				if tt.want.getPanic && r == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.getPanic && r != nil {
					t.Errorf("Unexpected panic")
				}
			}()

			ln := s.getListener(tt.args.addr)

			lnAddress := ln.Addr().String()
			if lnAddress != tt.want.addr {
				t.Errorf("Listener address: '%s', want '%s'", lnAddress, tt.want.addr)
			}

			lnNetwork := ln.Addr().Network()
			if lnNetwork != tt.want.network {
				t.Errorf("Listener network: '%s', want '%s'", lnNetwork, tt.want.network)
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
	testViewFn := func(ctx *fasthttp.RequestCtx) error {
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
				port:      8000,
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
				port:      8000,
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
				port:      8000,
				graceful:  true,
				tlsEnable: true,
			},
			want: want{
				getErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(&Config{
				Host:           "localhost",
				Port:           8000,
				LogLevel:       "error",
				TLSEnable:      tt.args.tlsEnable,
				GracefulEnable: tt.args.graceful,
			})

			serverCh := make(chan error, 1)
			go func() {
				err := s.ListenAndServe()
				serverCh <- err
			}()

			select {
			case err := <-serverCh:
				if !tt.want.getErr {
					t.Errorf("Unexpected error: %v", err)
				}
			case <-time.After(100 * time.Millisecond):
				if tt.want.getErr {
					t.Error("Error expected")
				}
			}
		})
	}
}
