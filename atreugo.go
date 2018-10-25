package atreugo

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

var allowedHTTPMethods = []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"}

// New create a new instance of Atreugo Server
func New(cfg *Config) *Atreugo {
	if cfg.Fasthttp == nil {
		cfg.Fasthttp = new(FasthttpConfig)
	}

	if cfg.Fasthttp.Name == "" {
		cfg.Fasthttp.Name = "AtreugoServer"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}
	if cfg.GracefulShutdown && cfg.Fasthttp.ReadTimeout <= 0 {
		cfg.Fasthttp.ReadTimeout = 20 * time.Second
	}

	r := router.New()

	handler := r.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	log := logger.New("atreugo", cfg.LogLevel, os.Stderr)

	server := &Atreugo{
		router: r,
		server: &fasthttp.Server{
			Handler:                       handler,
			Name:                          cfg.Fasthttp.Name,
			Concurrency:                   cfg.Fasthttp.Concurrency,
			DisableKeepalive:              cfg.Fasthttp.DisableKeepalive,
			ReadBufferSize:                cfg.Fasthttp.ReadBufferSize,
			WriteBufferSize:               cfg.Fasthttp.WriteBufferSize,
			ReadTimeout:                   cfg.Fasthttp.ReadTimeout,
			WriteTimeout:                  cfg.Fasthttp.WriteTimeout,
			MaxConnsPerIP:                 cfg.Fasthttp.MaxConnsPerIP,
			MaxRequestsPerConn:            cfg.Fasthttp.MaxRequestsPerConn,
			MaxKeepaliveDuration:          cfg.Fasthttp.MaxKeepaliveDuration,
			MaxRequestBodySize:            cfg.Fasthttp.MaxRequestBodySize,
			ReduceMemoryUsage:             cfg.Fasthttp.ReduceMemoryUsage,
			LogAllErrors:                  cfg.Fasthttp.LogAllErrors,
			DisableHeaderNamesNormalizing: cfg.Fasthttp.DisableHeaderNamesNormalizing,
			NoDefaultServerHeader:         cfg.Fasthttp.NoDefaultServerHeader,
			NoDefaultContentType:          cfg.Fasthttp.NoDefaultContentType,
			ConnState:                     cfg.Fasthttp.ConnState,
			Logger:                        log,
		},
		log: log,
		cfg: cfg,
	}

	return server
}

func (s *Atreugo) handler(viewFn View) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)
		defer releaseRequestCtx(actx)

		if s.log.DebugEnabled() {
			s.log.Debugf("%s %s", actx.Method(), actx.URI())
		}

		for _, middlewareFn := range s.middlewares {
			if statusCode, err := middlewareFn(actx); err != nil {
				s.log.Errorf("%s %s - %s", actx.Method(), actx.URI(), err)

				actx.Error(err.Error(), statusCode)
				return
			}
		}

		if err := viewFn(actx); err != nil {
			s.log.Error(err)
			actx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	}
}

// Serve serves incoming connections from the given listener.
//
// Serve blocks until the given listener returns permanent error.
func (s *Atreugo) Serve(ln net.Listener) error {
	schema := "http"
	if s.cfg.TLSEnable {
		schema = "https"
	}

	addr := ln.Addr().String()
	if addr != s.lnAddr {
		s.log.Info("Updating config with new listener address")
		sAddr := strings.Split(addr, ":")
		s.cfg.Host = sAddr[0]
		if len(sAddr) > 1 {
			s.cfg.Port, _ = strconv.Atoi(sAddr[1])
		} else {
			s.cfg.Port = 0
		}
		s.lnAddr = addr
	}

	s.log.Infof("Listening on: %s://%s/", schema, addr)
	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

func (s *Atreugo) serveGracefully(ln net.Listener) error {
	listenErr := make(chan error, 1)

	go func() {
		listenErr <- s.Serve(ln)
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		return err
	case <-osSignals:
		s.log.Infof("Shutdown signal received")

		if err := s.server.Shutdown(); err != nil {
			return err
		}

		s.log.Infof("Server gracefully stopped")
	}

	return nil
}

// Static add view for static files
func (s *Atreugo) Static(rootStaticDirPath string) {
	s.router.NotFound = fasthttp.FSHandler(rootStaticDirPath, 0)
}

// Path add the views to serve
func (s *Atreugo) Path(httpMethod string, url string, viewFn View) {
	if !include(allowedHTTPMethods, httpMethod) {
		panic("Invalid http method '" + httpMethod + "' for the url " + url)
	}

	s.router.Handle(httpMethod, url, s.handler(viewFn))
}

// UseMiddleware register middleware functions that viewHandler will use
func (s *Atreugo) UseMiddleware(fns ...Middleware) {
	s.middlewares = append(s.middlewares, fns...)
}

// SetLogOutput set log output of server
func (s *Atreugo) SetLogOutput(output io.Writer) {
	s.log.SetOutput(output)
}

// ListenAndServe serves HTTP/HTTPS requests from the given TCP4 addr.
//
// Pass custom listener to Serve if you need listening on non-TCP4 media
// such as IPv6.
func (s *Atreugo) ListenAndServe() error {
	s.lnAddr = fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := s.getListener()
	if err != nil {
		return err
	}

	if s.cfg.GracefulShutdown {
		return s.serveGracefully(ln)
	}

	return s.Serve(ln)
}
