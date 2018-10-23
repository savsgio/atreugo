package atreugo

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

var allowedHTTPMethods = []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"}

// New create a new instance of Atreugo Server
func New(cfg *Config) *Atreugo {
	if cfg.Name == "" {
		cfg.Name = "AtreugoServer"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}
	if cfg.GracefulShutdown && cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 20 * time.Second
	}

	r := router.New()

	handler := r.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	server := &Atreugo{
		router: r,
		server: &fasthttp.Server{
			Handler:     handler,
			Name:        cfg.Name,
			ReadTimeout: cfg.ReadTimeout,
		},
		log: logger.New("atreugo", cfg.LogLevel, os.Stdout),
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
				s.log.Errorf("Msg: %v | RequestUri: %s", err, actx.URI().String())

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

func (s *Atreugo) serve(ln net.Listener) error {
	schema := "http"
	if s.cfg.TLSEnable {
		schema = "https"
	}

	s.log.Infof("Listening on: %s://%s/", schema, ln.Addr().String())
	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

func (s *Atreugo) serveGracefully(ln net.Listener) error {
	listenErr := make(chan error, 1)

	go func() {
		listenErr <- s.serve(ln)
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

// ListenAndServe start Atreugo server according to the configuration
func (s *Atreugo) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln := s.getListener(addr)

	if s.cfg.GracefulShutdown {
		return s.serveGracefully(ln)
	}

	return s.serve(ln)
}
