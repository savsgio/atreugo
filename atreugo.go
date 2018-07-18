package atreugo

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/erikdubbelboer/fasthttp/reuseport"
	"github.com/savsgio/go-logger"
	"github.com/thehowl/fasthttprouter"
)

var allowedHTTPMethods = []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"}

// New create a new instance of Atreugo Server
func New(cfg *Config) *Atreugo {
	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}

	router := fasthttprouter.New()

	handler := router.Handler
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	server := &Atreugo{
		router: router,
		server: &fasthttp.Server{
			Handler:     handler,
			Name:        "AtreugoFastHTTPServer",
			ReadTimeout: 25 * time.Second,
		},
		log: logger.New("atreugo", cfg.LogLevel, os.Stdout),
		cfg: cfg,
	}

	return server
}

func (s *Atreugo) handler(viewFn View) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		s.log.Debugf("%s %s", ctx.Method(), ctx.URI())

		for _, middlewareFn := range s.middlewares {
			if statusCode, err := middlewareFn(ctx); err != nil {
				s.log.Errorf("Msg: %v | RequestUri: %s", err, ctx.URI().String())

				ctx.Error(err.Error(), statusCode)
				return
			}
		}

		if err := viewFn(ctx); err != nil {
			s.log.Error(err)
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	})
}

func (s *Atreugo) getListener(addr string) net.Listener {
	ln, err := reuseport.Listen(network, addr)
	if err == nil {
		return ln
	}
	s.log.Warningf("Error in reuseport listener %s", err)

	s.log.Infof("Trying with net listener")
	ln, err = net.Listen(network, addr)
	if err != nil {
		s.log.Fatalf("Error in net listener %s", err)
	}

	return ln
}

func (s *Atreugo) serve(ln net.Listener, protocol, addr string) error {
	s.log.Infof("Listening on: %s://%s/", protocol, addr)
	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

func (s *Atreugo) serveGracefully(ln net.Listener, protocol, addr string) error {
	listenErr := make(chan error, 1)

	go func() {
		listenErr <- s.serve(ln, protocol, addr)
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

// ListenAndServe start Atreugo server according to the configuration
func (s *Atreugo) ListenAndServe() error {
	protocol := "http"
	if s.cfg.TLSEnable {
		protocol = "https"
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln := s.getListener(addr)

	if s.cfg.GracefulEnable {
		return s.serveGracefully(ln, protocol, addr)
	}

	return s.serve(ln, protocol, addr)
}
