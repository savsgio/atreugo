package atreugo

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"time"

	logger "github.com/savsgio/go-logger"
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
		logLevel        string
		notFoundHandler fasthttp.RequestHandler
	}
	type want struct {
		logLevel        string
		notFoundHandler fasthttp.RequestHandler
	}

	notFoundHandler := func(ctx *fasthttp.RequestCtx) {}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Default",
			args: args{
				logLevel:        "",
				notFoundHandler: notFoundHandler,
			},
			want: want{
				logLevel:        logger.INFO,
				notFoundHandler: notFoundHandler,
			},
		},
		{
			name: "Custom",
			args: args{
				logLevel:        logger.WARNING,
				notFoundHandler: nil,
			},
			want: want{
				logLevel:        logger.WARNING,
				notFoundHandler: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.args.logLevel, NotFoundHandler: tt.args.notFoundHandler}
			s := New(cfg)

			if cfg.LogLevel != tt.want.logLevel {
				t.Errorf("Log level = %v, want %v", cfg.LogLevel, tt.want.logLevel)
			}

			if s.router == nil {
				t.Fatal("Atreugo router instance is nil")
			}

			if reflect.ValueOf(s.router.router.NotFound).Pointer() != reflect.ValueOf(tt.want.notFoundHandler).Pointer() {
				t.Errorf("Invalid NotFoundHandler = %p, want %p", s.router.router.NotFound, tt.want.notFoundHandler)
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

		if s.cfg.Fasthttp.ReadTimeout != defaultReadTimeout {
			t.Errorf("Config.Fasthttp.ReadTimeout = %v, want %v", s.cfg.Fasthttp.ReadTimeout, defaultReadTimeout)
		}

		if s.server.ReadTimeout != defaultReadTimeout {
			t.Errorf("fasthttp.Server.ReadTimeout = %v, want %v", s.server.ReadTimeout, defaultReadTimeout)
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
