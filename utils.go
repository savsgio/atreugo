package atreugo

import "github.com/valyala/fasthttp"

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
