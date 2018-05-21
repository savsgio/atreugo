package atreugo

import (
	"fmt"
	"os"

	"github.com/erikdubbelboer/fasthttp"
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
	log := logger.New("atreugo", logger.DEBUG, os.Stdout)

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
	addr := fmt.Sprintf("%s:%d", host, port)

	if len(logLevel) > 0 {
		server.log.SetLevel(logLevel[0])
	}

	server.log.Infof("Listening on: http://%s/", addr)
	server.log.Fatal(server.server.ListenAndServe(addr))
}
