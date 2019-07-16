package atreugo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	logger "github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var emptyFilters = Filters{}

// New create a new instance of Atreugo Server
func New(cfg *Config) *Atreugo {
	if cfg.Fasthttp == nil {
		cfg.Fasthttp = new(FasthttpConfig)
	}

	if cfg.Fasthttp.Name == "" {
		cfg.Fasthttp.Name = defaultServerName
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}
	if cfg.GracefulShutdown && cfg.Fasthttp.ReadTimeout <= 0 {
		cfg.Fasthttp.ReadTimeout = defaultReadTimeout
	}

	if cfg.LogName == "" {
		cfg.LogName = defaultLogName
	}

	r := router.New()
	if cfg.NotFoundHandler != nil {
		r.NotFound = cfg.NotFoundHandler
	}

	handler := r.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	log := logger.New(cfg.LogName, cfg.LogLevel, os.Stderr)

	server := &Atreugo{
		router: r,
		server: &fasthttp.Server{
			Handler:                            handler,
			Name:                               cfg.Fasthttp.Name,
			Concurrency:                        cfg.Fasthttp.Concurrency,
			DisableKeepalive:                   cfg.Fasthttp.DisableKeepalive,
			ReadBufferSize:                     cfg.Fasthttp.ReadBufferSize,
			WriteBufferSize:                    cfg.Fasthttp.WriteBufferSize,
			ReadTimeout:                        cfg.Fasthttp.ReadTimeout,
			WriteTimeout:                       cfg.Fasthttp.WriteTimeout,
			IdleTimeout:                        cfg.Fasthttp.IdleTimeout,
			MaxConnsPerIP:                      cfg.Fasthttp.MaxConnsPerIP,
			MaxRequestsPerConn:                 cfg.Fasthttp.MaxRequestsPerConn,
			MaxKeepaliveDuration:               cfg.Fasthttp.MaxKeepaliveDuration,
			MaxRequestBodySize:                 cfg.Fasthttp.MaxRequestBodySize,
			ReduceMemoryUsage:                  cfg.Fasthttp.ReduceMemoryUsage,
			GetOnly:                            cfg.Fasthttp.GetOnly,
			LogAllErrors:                       cfg.Fasthttp.LogAllErrors,
			DisableHeaderNamesNormalizing:      cfg.Fasthttp.DisableHeaderNamesNormalizing,
			SleepWhenConcurrencyLimitsExceeded: cfg.Fasthttp.SleepWhenConcurrencyLimitsExceeded,
			NoDefaultServerHeader:              cfg.Fasthttp.NoDefaultServerHeader,
			NoDefaultContentType:               cfg.Fasthttp.NoDefaultContentType,
			ConnState:                          cfg.Fasthttp.ConnState,
			KeepHijackedConns:                  cfg.Fasthttp.KeepHijackedConns,
			Logger:                             log,
		},
		log: log,
		cfg: cfg,
	}

	return server
}

func (s *Atreugo) handler(viewFn View, filters Filters) fasthttp.RequestHandler {
	before := append(s.beforeMiddlewares, filters.Before...)
	after := append(filters.After, s.afterMiddlewares...)

	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if s.log.DebugEnabled() {
			s.log.Debugf("%s %s", actx.Method(), actx.URI())
		}

		var err error
		var statusCode int

		if statusCode, err = execMiddlewares(actx, before); err == nil {
			if err = viewFn(actx); err != nil {
				statusCode = fasthttp.StatusInternalServerError
			} else {
				statusCode, err = execMiddlewares(actx, after)
			}
		}

		if err != nil {
			s.log.Error(err)
			actx.Error(err.Error(), statusCode)
		}

		releaseRequestCtx(actx)
	}
}

func (s *Atreugo) addRoute(httpMethod, url string, handler fasthttp.RequestHandler) {
	if httpMethod != strings.ToUpper(httpMethod) {
		panic("The http method '" + httpMethod + "' must be in uppercase")
	}

	s.router.Handle(httpMethod, url, handler)
}

// Serve serves incoming connections from the given listener.
//
// Serve blocks until the given listener returns permanent error.
//
// If use a custom Listener, will be updated your atreugo configuration
// with the Listener address automatically
func (s *Atreugo) Serve(ln net.Listener) error {
	schema := "http"
	if s.cfg.TLSEnable {
		schema = "https"
	}

	addr := ln.Addr().String()
	if addr != s.lnAddr {
		s.log.Info("Updating address config with the new listener address")
		sAddr := strings.Split(addr, ":")
		s.cfg.Host = sAddr[0]
		if len(sAddr) > 1 {
			s.cfg.Port, _ = strconv.Atoi(sAddr[1])
		} else {
			s.cfg.Port = 0
		}
		s.lnAddr = addr
	}

	s.log.Infof("Listening on: %s://%s/", schema, s.lnAddr)
	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

// ServeGracefully serves incoming connections from the given listener with graceful shutdown
//
// ServeGracefully blocks until the given listener returns permanent error.
//
// If use a custom Listener, will be updated your atreugo configuration
// with the Listener address and setting GracefulShutdown to true automatically.
func (s *Atreugo) ServeGracefully(ln net.Listener) error {
	if !s.cfg.GracefulShutdown {
		s.log.Info("Updating GracefulShutdown config to 'true'")
		s.cfg.GracefulShutdown = true

		if s.server.ReadTimeout <= 0 {
			s.log.Infof("Updating ReadTimeout config to '%v'", defaultReadTimeout)
			s.server.ReadTimeout = defaultReadTimeout
			s.cfg.Fasthttp.ReadTimeout = defaultReadTimeout
		}
	}

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

// Static serves static files from the given file system root.
func (s *Atreugo) Static(url, rootPath string) {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	s.router.ServeFiles(url+"/*filepath", rootPath)
}

// ServeFile serves a file from the given system path
func (s *Atreugo) ServeFile(url, filePath string) {
	s.Path("GET", url, func(ctx *RequestCtx) error {
		fasthttp.ServeFile(ctx.RequestCtx, filePath)
		return nil
	})
}

// Path add the view to serve from the given path and method
func (s *Atreugo) Path(httpMethod, url string, viewFn View) {
	s.PathWithFilters(httpMethod, url, viewFn, emptyFilters)
}

// PathWithFilters add the view to serve from the given path and method
// with filters that will execute before and after
func (s *Atreugo) PathWithFilters(httpMethod, url string, viewFn View, filters Filters) {
	s.addRoute(httpMethod, url, s.handler(viewFn, filters))
}

// TimeoutPath add the view to serve from the given path and method,
// which returns StatusRequestTimeout error with the given msg to the client
// if view didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view are running at the moment.
func (s *Atreugo) TimeoutPath(httpMethod, url string, viewFn View, timeout time.Duration, msg string) {
	s.TimeoutPathWithFilters(httpMethod, url, viewFn, emptyFilters, timeout, msg)
}

// TimeoutPathWithFilters add the view to serve from the given path and method
// with filters, which returns StatusRequestTimeout error with the given msg
// to the client if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (s *Atreugo) TimeoutPathWithFilters(httpMethod, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string) {
	handler := s.handler(viewFn, filters)
	s.addRoute(httpMethod, url, fasthttp.TimeoutHandler(handler, timeout, msg))
}

// TimeoutWithCodePath add the view to serve from the given path and method,
// which returns an error with the given msg and status code to the client
// if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (s *Atreugo) TimeoutWithCodePath(httpMethod, url string, viewFn View,
	timeout time.Duration, msg string, statusCode int) {
	s.TimeoutWithCodePathWithFilters(httpMethod, url, viewFn, emptyFilters, timeout, msg, statusCode)
}

// TimeoutWithCodePathWithFilters add the view to serve from the given path and method
// with filters, which returns an error with the given msg and status code to the client
// if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (s *Atreugo) TimeoutWithCodePathWithFilters(httpMethod, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string, statusCode int) {
	handler := s.handler(viewFn, filters)
	s.addRoute(httpMethod, url, fasthttp.TimeoutWithCodeHandler(handler, timeout, msg, statusCode))
}

// NetHTTPPath wraps net/http handler to atreugo view for the given path and method
//
// While this function may be used for easy switching from net/http to fasthttp/atreugo,
// it has the following drawbacks comparing to using manually written fasthttp/atreugo,
// request handler:
//
//     * A lot of useful functionality provided by fasthttp/atreugo is missing
//       from net/http handler.
//     * net/http -> fasthttp/atreugo handler conversion has some overhead,
//       so the returned handler will be always slower than manually written
//       fasthttp/atreugo handler.
//
// So it is advisable using this function only for quick net/http -> fasthttp
// switching. Then manually convert net/http handlers to fasthttp handlers
// according to https://github.com/valyala/fasthttp#switching-from-nethttp-to-fasthttp .
func (s *Atreugo) NetHTTPPath(httpMethod, url string, handler http.Handler) {
	s.NetHTTPPathWithFilters(httpMethod, url, handler, emptyFilters)
}

// NetHTTPPathWithFilters wraps net/http handler to atreugo view for the given path and method
// with filters that will execute before and after
//
// While this function may be used for easy switching from net/http to fasthttp/atreugo,
// it has the following drawbacks comparing to using manually written fasthttp/atreugo,
// request handler:
//
//     * A lot of useful functionality provided by fasthttp/atreugo is missing
//       from net/http handler.
//     * net/http -> fasthttp/atreugo handler conversion has some overhead,
//       so the returned handler will be always slower than manually written
//       fasthttp/atreugo handler.
//
// So it is advisable using this function only for quick net/http -> fasthttp
// switching. Then manually convert net/http handlers to fasthttp handlers
// according to https://github.com/valyala/fasthttp#switching-from-nethttp-to-fasthttp .
func (s *Atreugo) NetHTTPPathWithFilters(httpMethod, url string, handler http.Handler, filters Filters) {
	h := fasthttpadaptor.NewFastHTTPHandler(handler)

	aHandler := func(ctx *RequestCtx) error {
		h(ctx.RequestCtx)
		return nil
	}

	s.addRoute(httpMethod, url, s.handler(aHandler, filters))
}

// UseBefore register middleware functions in the order you want to execute them before the view execution
func (s *Atreugo) UseBefore(fns ...Middleware) {
	s.beforeMiddlewares = append(s.beforeMiddlewares, fns...)
}

// UseAfter register middleware functions in the order you want to execute them after the view execution
func (s *Atreugo) UseAfter(fns ...Middleware) {
	s.afterMiddlewares = append(s.afterMiddlewares, fns...)
}

// SetLogOutput set log output of server
func (s *Atreugo) SetLogOutput(output io.Writer) {
	s.log.SetOutput(output)
}

// ListenAndServe serves HTTP/HTTPS requests from the given TCP4 addr in the atreugo configuration.
//
// Pass custom listener to Serve/ServeGracefully if you need listening on non-TCP4 media
// such as IPv6.
func (s *Atreugo) ListenAndServe() error {
	s.lnAddr = fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := s.getListener()
	if err != nil {
		return err
	}

	if s.cfg.GracefulShutdown {
		return s.ServeGracefully(ln)
	}

	return s.Serve(ln)
}
