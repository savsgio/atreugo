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

func defaultErrorView(ctx *RequestCtx, err error, statusCode int) {
	ctx.Error(err.Error(), statusCode)
}

func newRouter(log *logger.Logger, errorView ErrorView) *Router {
	r := new(Router)
	r.router = fastrouter.New()
	r.beginPath = "/"
	r.log = log

	if errorView == nil {
		errorView = defaultErrorView
	}

	r.errorView = errorView

	return r
}

// NewGroupPath returns a new router to group paths
func (r *Router) NewGroupPath(path string) *Router {
	g := new(Router)
	g.router = r.router.Group(path)
	g.parent = r
	g.beginPath = path
	g.errorView = r.errorView
	g.log = r.log

	return g
}

func (r *Router) init() {
	if r.parent != nil {
		panic("Could not be executed by group router")
	}

	for _, p := range r.paths {
		handler := p.handlerBuilder(p.view, p.middlewares)
		if p.withTimeout {
			handler = fasthttp.TimeoutWithCodeHandler(handler, p.timeout, p.timeoutMsg, p.timeoutCode)
		}

		r.router.Handle(p.method, p.url, handler)
	}
}

func (r *Router) buildMiddlewaresChain(skip ...Middleware) Middlewares {
	mdlws := Middlewares{}

	var subMdlws Middlewares

	if r.parent != nil {
		skip = append(skip, r.middlewares.Skip...)
		subMdlws = r.parent.buildMiddlewaresChain(skip...)
	}

	mdlws.Before = appendMiddlewares(mdlws.Before, subMdlws.Before, skip...)
	mdlws.Before = appendMiddlewares(mdlws.Before, r.middlewares.Before, skip...)

	mdlws.After = appendMiddlewares(mdlws.After, r.middlewares.After, skip...)
	mdlws.After = appendMiddlewares(mdlws.After, subMdlws.After, skip...)

	return mdlws
}

func (r *Router) getGroupFullPath(path string) string {
	if r.beginPath != "/" {
		path = r.beginPath + path
	}

	if r.parent != nil {
		path = r.parent.getGroupFullPath(path)
	}

	return path
}

func (r *Router) appendPath(p *Path) {
	if r.parent != nil {
		r.parent.appendPath(p)
		return
	}

	r.paths = append(r.paths, p)
}

func (r *Router) addPath(method, url string, fn View) *Path {
	if method != strings.ToUpper(method) {
		panic("The http method '" + method + "' must be in uppercase")
	}

	p := &Path{handlerBuilder: r.handler, method: method, url: r.getGroupFullPath(url), view: fn}
	r.appendPath(p)

	return p
}

func (r *Router) handler(fn View, middle Middlewares) fasthttp.RequestHandler {
	mdlws := r.buildMiddlewaresChain(middle.Skip...)

	chain := append(mdlws.Before, middle.Before...)
	chain = append(chain, func(ctx *RequestCtx) error {
		if !ctx.skipView {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return ctx.Next()
	})
	chain = append(chain, middle.After...)
	chain = append(chain, mdlws.After...)

	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if r.log.DebugEnabled() {
			r.log.Debugf("%s %s", actx.Method(), actx.URI())
		}

		if err := execute(actx, chain); err != nil {
			statusCode := actx.Response.StatusCode()
			if statusCode == fasthttp.StatusOK {
				statusCode = fasthttp.StatusInternalServerError
			}

			r.log.Error(err)
			r.errorView(actx, err, statusCode)
		}

		releaseRequestCtx(actx)
	}
}

// Middlewares defines the middlewares (before, after and skip) in the order in which you want to execute them
// for the view or group
//
// ** The previous middlewares configuration could be overridden
func (r *Router) Middlewares(middlewares Middlewares) *Router {
	r.middlewares = middlewares

	return r
}

// UseBefore registers the middlewares in the order in which you want to execute them
// before the execution of the view or group
func (r *Router) UseBefore(fns ...Middleware) *Router {
	r.middlewares.Before = append(r.middlewares.Before, fns...)

	return r
}

// UseAfter registers the middlewares in the order in which you want to execute them
// after the execution of the view or group
func (r *Router) UseAfter(fns ...Middleware) *Router {
	r.middlewares.After = append(r.middlewares.After, fns...)

	return r
}

// SkipMiddlewares registers the middlewares that you want to skip when executing the view or group
func (r *Router) SkipMiddlewares(fns ...Middleware) *Router {
	r.middlewares.Skip = append(r.middlewares.Skip, fns...)

	return r
}

// GET shortcut for router.Path("GET", url, viewFn)
func (r *Router) GET(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodGet, url, viewFn)
}

// HEAD shortcut for router.Path("HEAD", url, viewFn)
func (r *Router) HEAD(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodHead, url, viewFn)
}

// OPTIONS shortcut for router.Path("OPTIONS", url, viewFn)
func (r *Router) OPTIONS(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodOptions, url, viewFn)
}

// POST shortcut for router.Path("POST", url, viewFn)
func (r *Router) POST(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodPost, url, viewFn)
}

// PUT shortcut for router.Path("PUT", url, viewFn)
func (r *Router) PUT(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodPut, url, viewFn)
}

// PATCH shortcut for router.Path("PATCH", url, viewFn)
func (r *Router) PATCH(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodPatch, url, viewFn)
}

// DELETE shortcut for router.Path("DELETE", url, viewFn)
func (r *Router) DELETE(url string, viewFn View) *Path {
	return r.Path(fasthttp.MethodDelete, url, viewFn)
}

// Path registers a new view with the given path and method
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy)
func (r *Router) Path(method, url string, viewFn View) *Path {
	return r.addPath(method, url, viewFn)
}

// RequestHandlerPath wraps fasthttp request handler to atreugo view and registers it to
// the given path and method
func (r *Router) RequestHandlerPath(method, url string, handler fasthttp.RequestHandler) *Path {
	viewFn := func(ctx *RequestCtx) error {
		handler(ctx.RequestCtx)
		return nil
	}

	return r.addPath(method, url, viewFn)
}

// NetHTTPPath wraps net/http handler to atreugo view and registers it to
// the given path and method.
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
// according to https://github.com/valyala/fasthttp#switching-from-nethttp-to-fasthttp
func (r *Router) NetHTTPPath(method, url string, handler http.Handler) *Path {
	h := fasthttpadaptor.NewFastHTTPHandler(handler)

	return r.RequestHandlerPath(method, url, h)
}

// Static serves static files from the given file system root
//
// Make sure your program has enough 'max open files' limit aka
// 'ulimit -n' if root folder contains many files
func (r *Router) Static(url, rootPath string) *Path {
	return r.StaticCustom(url, &StaticFS{
		Root:               rootPath,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		AcceptByteRange:    true,
	})
}

// StaticCustom serves static files from the given file system settings
//
// Make sure your program has enough 'max open files' limit aka
// 'ulimit -n' if root folder contains many files
func (r *Router) StaticCustom(url string, fs *StaticFS) *Path {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	ffs := &fasthttp.FS{
		Root:                 fs.Root,
		IndexNames:           fs.IndexNames,
		GenerateIndexPages:   fs.GenerateIndexPages,
		Compress:             fs.Compress,
		AcceptByteRange:      fs.AcceptByteRange,
		CacheDuration:        fs.CacheDuration,
		CompressedFileSuffix: fs.CompressedFileSuffix,
	}

	if fs.PathNotFound != nil {
		ffs.PathNotFound = viewToHandler(fs.PathNotFound, r.errorView)
	}

	if fs.PathRewrite != nil {
		ffs.PathRewrite = func(ctx *fasthttp.RequestCtx) []byte {
			actx := acquireRequestCtx(ctx)
			result := fs.PathRewrite(actx)
			releaseRequestCtx(actx)

			return result
		}
	}

	stripSlashes := strings.Count(r.getGroupFullPath(url), "/")

	if ffs.PathRewrite == nil && stripSlashes > 0 {
		ffs.PathRewrite = fasthttp.NewPathSlashesStripper(stripSlashes)
	}

	p := r.RequestHandlerPath(fasthttp.MethodGet, url+"/*filepath", ffs.NewRequestHandler())

	// nolint:godox
	// TODO: Remove in version v11.0.0
	if len(fs.Filters.Before) > 0 || len(fs.Filters.After) > 0 || len(fs.Filters.Skip) > 0 {
		p.Middlewares(Middlewares(fs.Filters))
	}

	return p
}

// ServeFile returns HTTP response containing compressed file contents
// from the given path
//
// HTTP response may contain uncompressed file contents in the following cases:
//
//   * Missing 'Accept-Encoding: gzip' request header.
//   * No write access to directory containing the file.
//
// Directory contents is returned if path points to directory
func (r *Router) ServeFile(url, filePath string) *Path {
	viewFn := func(ctx *RequestCtx) error {
		fasthttp.ServeFile(ctx.RequestCtx, filePath)
		return nil
	}

	return r.addPath(fasthttp.MethodGet, url, viewFn)
}

// ListPaths returns all registered routes grouped by method
func (r *Router) ListPaths() map[string][]string {
	return r.router.List()
}

//
// DEPRECATED
//

// PathWithFilters registers a new view with the given path and method,
// and with filters that will execute before and after.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Path(method, url, viewFn).Middlewares(middlewares)
func (r *Router) PathWithFilters(method, url string, viewFn View, filters Filters) {
	r.Path(method, url, viewFn).Middlewares(Middlewares(filters))
}

// RequestHandlerPathWithFilters wraps fasthttp request handler to atreugo view and registers it to
// the given path and method, and with filters that will execute before and after.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.RequestHandlerPath(method, url, handler).Middlewares(middlewares)
func (r *Router) RequestHandlerPathWithFilters(method, url string, handler fasthttp.RequestHandler,
	filters Filters) {
	r.RequestHandlerPath(method, url, handler).Middlewares(Middlewares(filters))
}

// TimeoutPath registers a new view with the given path and method,
// which returns StatusRequestTimeout error with the given msg to the client
// if view didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view are running at the moment.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Path(method, url, viewFn).Timeout(timeout, msg)
func (r *Router) TimeoutPath(method, url string, viewFn View, timeout time.Duration, msg string) {
	r.TimeoutPathWithFilters(method, url, viewFn, Filters{}, timeout, msg)
}

// TimeoutPathWithFilters registers a new view with the given path and method,
// and with filters that will execute before and after, which returns StatusRequestTimeout
// error with the given msg to the client if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Path(method, url, viewFn).Timeout(timeout, msg).Middlewares(middlewares)
func (r *Router) TimeoutPathWithFilters(method, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string) {
	r.Path(method, url, viewFn).Timeout(timeout, msg).Middlewares(Middlewares(filters))
}

// TimeoutWithCodePath registers a new view with the given path and method,
// which returns an error with the given msg and status code to the client
// if view/filters didn't return during the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Path(method, url, viewFn).TimeoutCode(timeout, msg, statusCode)
func (r *Router) TimeoutWithCodePath(method, url string, viewFn View,
	timeout time.Duration, msg string, statusCode int) {
	r.TimeoutWithCodePathWithFilters(method, url, viewFn, Filters{}, timeout, msg, statusCode)
}

// TimeoutWithCodePathWithFilters registers a new view with the given path and method,
// and with filters that will execute before and after, which returns an error
// with the given msg and status code to the client if view/filters didn't return during
// the given duration.
//
// The returned handler may return StatusTooManyRequests error with the given
// msg to the client if there are more than Server.Concurrency concurrent
// handlers view/filters are running at the moment.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Path(method, url, viewFn).TimeoutCode(timeout, msg, statusCode).Middlewares(middlewares)
func (r *Router) TimeoutWithCodePathWithFilters(method, url string, viewFn View, filters Filters,
	timeout time.Duration, msg string, statusCode int) {
	r.Path(method, url, viewFn).TimeoutCode(timeout, msg, statusCode).Middlewares(Middlewares(filters))
}

// NetHTTPPathWithFilters wraps net/http handler to atreugo view and registers it to
// the given path and method, and with filters that will execute before and after
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
// according to https://github.com/valyala/fasthttp#switching-from-nethttp-to-fasthttp.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.NetHTTPPath(method, url, handler).Middlewares(middlewares)
func (r *Router) NetHTTPPathWithFilters(method, url string, handler http.Handler, filters Filters) {
	r.NetHTTPPath(method, url, handler).Middlewares(Middlewares(filters))
}

// StaticWithFilters serves static files from the given file system root,
// and with filters that will execute before and after request a file.
//
// Make sure your program has enough 'max open files' limit aka
// 'ulimit -n' if root folder contains many files.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.Static(url, rootPath).Middlewares(middlewares)
func (r *Router) StaticWithFilters(url, rootPath string, filters Filters) {
	r.Static(url, rootPath).Middlewares(Middlewares(filters))
}

// ServeFileWithFilters returns HTTP response containing compressed file contents
// from the given path, and with filters that will execute before and after request the file.
//
// HTTP response may contain uncompressed file contents in the following cases:
//
//   * Missing 'Accept-Encoding: gzip' request header.
//   * No write access to directory containing the file.
//
// Directory contents is returned if path points to directory.
//
// WARNING: It's deprecated, will be remove in version v11.0.0.
// Use instead:
// 		r.ServeFile(url, filePath).Middlewares(middlewares)
func (r *Router) ServeFileWithFilters(url, filePath string, filters Filters) {
	r.ServeFile(url, filePath).Middlewares(Middlewares(filters))
}
