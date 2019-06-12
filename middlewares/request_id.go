package middlewares

import (
	"github.com/google/uuid"
	"github.com/savsgio/atreugo/v7"
	"github.com/valyala/fasthttp"
)

// RequestIDMiddleware adds an identifier to the request
// Header name: X-Request-ID
// Header value: uuid4()
//
// It's recomemded to add this middleware at first and before the view execution.
func RequestIDMiddleware(ctx *atreugo.RequestCtx) (int, error) {
	if ctx.Request.Header.Peek(atreugo.XRequestIDHeader) == nil {
		ctx.Request.Header.Set(atreugo.XRequestIDHeader, uuid.New().String())
	}

	return fasthttp.StatusOK, nil
}
