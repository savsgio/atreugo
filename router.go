package atreugo

import (
	"net/http"
	"strings"

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

	return r.RequestHandlerPath(fasthttp.MethodGet, url+"/*filepath", ffs.NewRequestHandler())
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
