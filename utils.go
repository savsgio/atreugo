package atreugo

import "github.com/valyala/fasthttp"

// index returns the first index of the target string `t`, or
// -1 if no match is found.
func indexOf(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// include returns `true` if the target string t is in the
// slice.
func include(vs []string, t string) bool {
	return indexOf(vs, t) >= 0
}

// execMiddlewares execute all the middlewares functions with the request context given
func execMiddlewares(ctx *RequestCtx, middlewares []Middleware) (int, error) {
	for _, middlewareFn := range middlewares {
		if statusCode, err := middlewareFn(ctx); statusCode != 0 || err != nil {
			return statusCode, err
		}
	}

	return 0, nil
}

func viewToHandler(view View) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if err := view(actx); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		releaseRequestCtx(actx)
	}
}
