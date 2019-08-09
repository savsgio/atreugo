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

	logger "github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

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
	if cfg.Network == "" {
		cfg.Network = defaultNetwork
	}
	if cfg.GracefulShutdown && cfg.Fasthttp.ReadTimeout <= 0 {
		cfg.Fasthttp.ReadTimeout = defaultReadTimeout
	}

	if cfg.LogName == "" {
		cfg.LogName = defaultLogName
	}

	log := logger.New(cfg.LogName, cfg.LogLevel, os.Stderr)

	server := &Atreugo{
		server: &fasthttp.Server{
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

	r := newRouter(log)
	if cfg.NotFoundHandler != nil {
		r.router.NotFound = cfg.NotFoundHandler
	}

	handler := r.router.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	server.Router = r
	server.server.Handler = handler

	return server
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
