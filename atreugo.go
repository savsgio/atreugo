package atreugo

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/erikdubbelboer/fasthttp/reuseport"
	"github.com/savsgio/go-logger"
	"github.com/thehowl/fasthttprouter"
)

// Atreugo config for make up a server
type Atreugo struct {
	server      *fasthttp.Server
	router      *fasthttprouter.Router
	middlewares []middleware
	log         *logger.Logger
}

type view func(ctx *fasthttp.RequestCtx) error
type middleware func(ctx *fasthttp.RequestCtx) (int, error)

// New create a new instance of Atreugo Server
func New() *Atreugo {
	log := logger.New("atreugo", logger.INFO, os.Stdout)

	router := fasthttprouter.New()

	server := &Atreugo{
		router: router,
		server: &fasthttp.Server{
			Handler:     router.Handler,
			Name:        "AtreugoFastHTTPServer",
			ReadTimeout: 5 * time.Second,
		},
		log: log,
	}

	return server
}

func (s *Atreugo) viewHandler(viewFn view) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		s.log.Debugf("%s %s", ctx.Method(), ctx.URI())

		for _, middlewareFn := range s.middlewares {
			if statusCode, err := middlewareFn(ctx); err != nil {
				s.log.Errorf("Msg: %v | RequestUri: %s", err, ctx.URI().String())

				JsonResponse(ctx, Json{"Error": err.Error()}, statusCode)
				return
			}
		}

		if err := viewFn(ctx); err != nil {
			s.log.Error(err)
		}
	})
}

// Static add view for static files
func (s *Atreugo) Static(rootStaticDirPath string) {
	s.router.NotFound = fasthttp.FSHandler(rootStaticDirPath, 0)
}

// Path add the views to serve
func (s *Atreugo) Path(httpMethod string, url string, viewFn view) {
	callFuncByName(s.router, httpMethod, url, s.viewHandler(viewFn))
}

// UseMiddleware register middleware functions that viewHandler will use
func (s *Atreugo) UseMiddleware(fns ...middleware) {
	s.middlewares = append(s.middlewares, fns...)
}

// ListenAndServe start Atreugo server
func (s *Atreugo) ListenAndServe(host string, port int, logLevel ...string) {
	if len(logLevel) > 0 {
		s.log.SetLevel(logLevel[0])
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	ln, err := reuseport.Listen("tcp4", addr)
	if err != nil {
		s.log.Fatalf("Error in reuseport listener: %s", err)
	}

	// Error handling
	listenErr := make(chan error, 1)

	go func() {
		s.log.Infof("Listening on: http://%s/", addr)
		listenErr <- s.server.Serve(ln)
	}()

	// SIGINT/SIGTERM handling
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Handle channels/graceful shutdown
	select {
	// If s.Serve(ln) cannot start due to errors
	case err := <-listenErr:
		if err != nil {
			s.log.Fatalf("listener error: %s", err)
		}
	// handle termination signal
	case <-osSignals:
		s.log.Infof("Shutdown signal received")

		// Attempt the graceful shutdown by closing the listener
		// and completing all inflight requests.
		if err := s.server.Shutdown(); err != nil {
			s.log.Fatalf("unexepcted error: %s", err)
		}

		s.log.Infof("Server gracefully stopped")
	}
}
