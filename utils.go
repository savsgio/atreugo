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

// execute executes all middlewares + filters + view with the given request context
func execute(ctx *RequestCtx, hs []Middleware) error {
	for _, h := range hs {
		if err := h(ctx); err != nil {
			return err
		}
		if !ctx.next {
			return nil
		}
		ctx.next = false
	}

	return nil
}

func viewToHandler(view View, errorView ErrorView) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if err := view(actx); err != nil {
			errorView(actx, err, fasthttp.StatusInternalServerError)
		}

		releaseRequestCtx(actx)
	}
}
