package atreugo

import (
	"net/http"
	"strings"
	"time"

	fastrouter "github.com/fasthttp/router"
	logger "github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var emptyFilters = Filters{}

func newRouter(log *logger.Logger) *router {
	r := new(router)
	r.log = log
	r.router = fastrouter.New()

	return r
}

func (r *router) NewGroupPath(path string) *router {
	g := new(router)
	g.log = r.log
	g.router = r.router.Group(path)

	return g
}

func (r *router) addRoute(httpMethod, url string, handler fasthttp.RequestHandler) {
	if httpMethod != strings.ToUpper(httpMethod) {
		panic("The http method '" + httpMethod + "' must be in uppercase")
	}

	r.router.Handle(httpMethod, url, handler)
}

func (r *router) handler(viewFn View, filters Filters) fasthttp.RequestHandler {
	before := append(r.beforeMiddlewares, filters.Before...)
	after := append(filters.After, r.afterMiddlewares...)

	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if r.log.DebugEnabled() {
			r.log.Debugf("%s %s", actx.Method(), actx.URI())
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
			r.log.Error(err)
			actx.Error(err.Error(), statusCode)
		}

		releaseRequestCtx(actx)
	}
}

// UseBefore register middleware functions in the order you want to execute them before the view execution
func (r *router) UseBefore(fns ...Middleware) {
	r.beforeMiddlewares = append(r.beforeMiddlewares, fns...)
}

// UseAfter register middleware functions in the order you want to execute them after the view execution
func (r *router) UseAfter(fns ...Middleware) {
	r.afterMiddlewares = append(r.afterMiddlewares, fns...)
}

// Path add the view to serve from the given path and method
func (r *router) Path(httpMethod, url string, viewFn View) {
	r.PathWithFilters(httpMethod, url, viewFn, emptyFilters)
}

// PathWithFilters add the view to serve from the given path and method
// with filters that will execute before and after
func (r *router) PathWithFilters(httpMethod, url string, viewFn View, filters Filters) {
	r.addRoute(httpMethod, url, r.handler(viewFn, filters))
}

// TimeoutPath add the view to serve from the given path and method,
// which returns StatusRequestTimeout error with the given msg to the client
// if view didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view are running at the moment.
func (r *router) TimeoutPath(httpMethod, url string, viewFn View, timeout time.Duration, msg string) {
	r.TimeoutPathWithFilters(httpMethod, url, viewFn, emptyFilters, timeout, msg)
}

// TimeoutPathWithFilters add the view to serve from the given path and method
// with filters, which returns StatusRequestTimeout error with the given msg
// to the client if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (r *router) TimeoutPathWithFilters(httpMethod, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string) {
	handler := r.handler(viewFn, filters)
	r.addRoute(httpMethod, url, fasthttp.TimeoutHandler(handler, timeout, msg))
}

// TimeoutWithCodePath add the view to serve from the given path and method,
// which returns an error with the given msg and status code to the client
// if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (r *router) TimeoutWithCodePath(httpMethod, url string, viewFn View,
	timeout time.Duration, msg string, statusCode int) {
	r.TimeoutWithCodePathWithFilters(httpMethod, url, viewFn, emptyFilters, timeout, msg, statusCode)
}

// TimeoutWithCodePathWithFilters add the view to serve from the given path and method
// with filters, which returns an error with the given msg and status code to the client
// if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
func (r *router) TimeoutWithCodePathWithFilters(httpMethod, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string, statusCode int) {
	handler := r.handler(viewFn, filters)
	r.addRoute(httpMethod, url, fasthttp.TimeoutWithCodeHandler(handler, timeout, msg, statusCode))
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
func (r *router) NetHTTPPath(httpMethod, url string, handler http.Handler) {
	r.NetHTTPPathWithFilters(httpMethod, url, handler, emptyFilters)
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
func (r *router) NetHTTPPathWithFilters(httpMethod, url string, handler http.Handler, filters Filters) {
	h := fasthttpadaptor.NewFastHTTPHandler(handler)

	aHandler := func(ctx *RequestCtx) error {
		h(ctx.RequestCtx)
		return nil
	}

	r.addRoute(httpMethod, url, r.handler(aHandler, filters))
}

// Static serves static files from the given file system root.
func (r *router) Static(url, rootPath string) {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	r.router.ServeFiles(url+"/*filepath", rootPath)
}

// ServeFile serves a file from the given system path
func (r *router) ServeFile(url, filePath string) {
	r.router.GET(url, func(ctx *fasthttp.RequestCtx) {
		fasthttp.ServeFile(ctx, filePath)
	})
}
