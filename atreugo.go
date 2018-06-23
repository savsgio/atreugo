package atreugo

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
		router:      router,
		middlewares: []middleware{},
		server: &fasthttp.Server{
			Handler: router.Handler,
			Name:    "AtreugoFastHTTPServer",
		},
		log: log,
	}

	return server
}

func (server *Atreugo) viewHandler(viewFn view) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		server.log.Debugf("%s %s", ctx.Method(), ctx.URI())

		for _, middlewareFn := range server.middlewares {
			if statusCode, err := middlewareFn(ctx); err != nil {
				server.log.Errorf("Msg: %v | RequestUri: %s", err, ctx.URI().String())

				JsonResponse(ctx, Json{"Error": err.Error()}, statusCode)
				return
			}
		}

		if err := viewFn(ctx); err != nil {
			server.log.Error(err)
		}
	})
}

// Static add view for static files
func (server *Atreugo) Static(rootStaticDirPath string) {
	server.router.NotFound = fasthttp.FSHandler(rootStaticDirPath, 0)
}

// Path add the views to serve
func (server *Atreugo) Path(httpMethod string, url string, viewFn view) {
	callFuncByName(server.router, httpMethod, url, server.viewHandler(viewFn))
}

// UseMiddleware register middleware functions that viewHandler will use
func (server *Atreugo) UseMiddleware(fns ...middleware) {
	server.middlewares = append(server.middlewares, fns...)
}

// ListenAndServe start Atreugo server
func (server *Atreugo) ListenAndServe(host string, port int, logLevel ...string) {
	if len(logLevel) > 0 {
		server.log.SetLevel(logLevel[0])
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	ln, err := reuseport.Listen("tcp4", addr)
	if err != nil {
		server.log.Fatalf("Error in reuseport listener: %s", err)
	}

	// Error handling
	listenErr := make(chan error, 1)

	go func() {
		server.log.Infof("Listening on: http://%s/", addr)
		listenErr <- server.server.Serve(ln)
	}()

	// SIGINT/SIGTERM handling
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Handle channels/graceful shutdown
	for {
		select {
		// If server.ListenAndServe() cannot start due to errors such
		// as "port in use" it will return an error.
		case err := <-listenErr:
			if err != nil {
				server.log.Fatalf("listener error: %s", err)
			}
			os.Exit(0)
		// handle termination signal
		case <-osSignals:
			server.log.Infof("Shutdown signal received")

			// Servers in the process of shutting down should disable KeepAlives
			// FIXME: This causes a data race
			server.server.DisableKeepalive = true

			// Attempt the graceful shutdown by closing the listener
			// and completing all inflight requests.
			if err := server.server.Shutdown(); err != nil {
				server.log.Fatalf("unexepcted error: %s", err)
			}

			server.log.Infof("Server gracefully stopped")
		}
	}
}
