package atreugo

import (
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/savsgio/go-logger"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
)

var tcpNetworks = []string{"tcp", "tcp4", "tcp6"}
var validNetworks = append(tcpNetworks, []string{"unix"}...)

// New create a new instance of Atreugo Server
func New(cfg *Config) *Atreugo {
	if cfg.Network != "" && !gotils.StringSliceInclude(validNetworks, cfg.Network) {
		panic("Invalid network: " + cfg.Network)
	}
	if cfg.Network == "" {
		cfg.Network = defaultNetwork
	}

	if cfg.Name == "" {
		cfg.Name = defaultServerName
	}
	if cfg.LogName == "" {
		cfg.LogName = defaultLogName
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}
	if cfg.GracefulShutdown && cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}

	log := logger.New(cfg.LogName, cfg.LogLevel, os.Stderr)

	r := newRouter(log, cfg.ErrorView)

	if cfg.NotFoundView != nil {
		r.router.NotFound = viewToHandler(cfg.NotFoundView, r.errorView)
	}
	if cfg.MethodNotAllowedView != nil {
		r.router.MethodNotAllowed = viewToHandler(cfg.MethodNotAllowedView, r.errorView)
	}
	if cfg.PanicView != nil {
		r.router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
			actx := acquireRequestCtx(ctx)
			cfg.PanicView(actx, err)
			releaseRequestCtx(actx)
		}
	}

	handler := r.router.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	server := &Atreugo{
		server: &fasthttp.Server{
			Name:                               cfg.Name,
			Handler:                            handler,
			Concurrency:                        cfg.Concurrency,
			DisableKeepalive:                   cfg.DisableKeepalive,
			ReadBufferSize:                     cfg.ReadBufferSize,
			WriteBufferSize:                    cfg.WriteBufferSize,
			ReadTimeout:                        cfg.ReadTimeout,
			WriteTimeout:                       cfg.WriteTimeout,
			IdleTimeout:                        cfg.IdleTimeout,
			MaxConnsPerIP:                      cfg.MaxConnsPerIP,
			MaxRequestsPerConn:                 cfg.MaxRequestsPerConn,
			MaxKeepaliveDuration:               cfg.MaxKeepaliveDuration,
			MaxRequestBodySize:                 cfg.MaxRequestBodySize,
			ReduceMemoryUsage:                  cfg.ReduceMemoryUsage,
			GetOnly:                            cfg.GetOnly,
			LogAllErrors:                       cfg.LogAllErrors,
			DisableHeaderNamesNormalizing:      cfg.DisableHeaderNamesNormalizing,
			SleepWhenConcurrencyLimitsExceeded: cfg.SleepWhenConcurrencyLimitsExceeded,
			NoDefaultServerHeader:              cfg.NoDefaultServerHeader,
			NoDefaultContentType:               cfg.NoDefaultContentType,
			ConnState:                          cfg.ConnState,
			KeepHijackedConns:                  cfg.KeepHijackedConns,
			Logger:                             log,
		},

		log: log,
		cfg: cfg,

		Router: r,
	}

	return server
}

// Serve serves incoming connections from the given listener.
//
// Serve blocks until the given listener returns permanent error.
//
// If use a custom Listener, will be updated your atreugo configuration
// with the Listener address automatically
func (s *Atreugo) Serve(ln net.Listener) error {
	defer ln.Close()

	s.cfg.Addr = ln.Addr().String()
	s.cfg.Network = ln.Addr().Network()

	if gotils.StringSliceInclude(tcpNetworks, s.cfg.Network) {
		schema := "http"
		if s.cfg.TLSEnable {
			schema = "https"
		}

		s.log.Infof("Listening on: %s://%s/", schema, s.cfg.Addr)

	} else {
		s.log.Infof("Listening on (network: %s): %s ", s.cfg.Network, s.cfg.Addr)
	}

	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

// ServeGracefully serves incoming connections from the given listener with graceful shutdown
//
// ServeGracefully blocks until the given listener returns permanent error.
func (s *Atreugo) ServeGracefully(ln net.Listener) error {
	s.cfg.GracefulShutdown = true

	if s.server.ReadTimeout <= 0 {
		s.server.ReadTimeout = defaultReadTimeout
		s.cfg.ReadTimeout = defaultReadTimeout
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

// ListenAndServe serves requests from the given network and address in the atreugo configuration.
//
// Pass custom listener to Serve/ServeGracefully if you want to use it.
func (s *Atreugo) ListenAndServe() error {
	ln, err := s.getListener()
	if err != nil {
		return err
	}

	if s.cfg.GracefulShutdown {
		return s.ServeGracefully(ln)
	}

	return s.Serve(ln)
}
