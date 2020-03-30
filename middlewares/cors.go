package middlewares

import (
	"strconv"
	"strings"

	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
)

const strHeaderDelim = ", "

const allowOriginHeaderName = "Access-Control-Allow-Origin"
const allowCredentialsHeaderName = "Access-Control-Allow-Credentials"
const allowHeadersHeaderName = "Access-Control-Allow-Headers"
const allowMethodsHeaderName = "Access-Control-Allow-Methods"
const exposeHeadersHeaderName = "Access-Control-Expose-Headers"
const maxAgeHeaderName = "Access-Control-Max-Age"

const varyHeaderName = "Vary"
const originHeaderName = "Origin"

// CORS properties.
type CORS struct {
	// Specifies either the origins, which tells browsers to allow that origin
	// to access the resource; or else — for requests without credentials —
	// the "*" wildcard, to tell browsers to allow any origin to access the resource.
	AllowedOrigins []string

	// Specifies the method or methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	// The conditions under which a request is preflighted are discussed above.
	AllowedMethods []string

	// This is used in response to a preflight request to indicate which HTTP headers
	// can be used when making the actual request.
	AllowedHeaders []string

	// Indicates whether or not the response to the request can be exposed when
	// the credentials flag is true. When used as part of a response to a preflight request,
	// this indicates whether or not the actual request can be made using credentials.
	// Note that simple GET requests are not preflighted, and so if a request is made
	// for a resource with credentials, if this header is not returned with the resource,
	// the response is ignored by the browser and not returned to web content.
	AllowCredentials bool

	// Indicates how long, in seconds, the results of a preflight request can be cached
	AllowMaxAge int

	// Header or headers to lets a server whitelist headers that browsers are allowed to access.
	ExposedHeaders []string
}

// NewCORSMiddleware returns the middleware with the configured properties
//
// IMPORTANT: always use as last middleware (`server.UseAfter(...)`)
func NewCORSMiddleware(cors CORS) atreugo.Middleware {
	allowedHeaders := strings.Join(cors.AllowedHeaders, strHeaderDelim)
	allowedMethods := strings.Join(cors.AllowedMethods, strHeaderDelim)
	exposedHeaders := strings.Join(cors.ExposedHeaders, strHeaderDelim)
	maxAge := strconv.Itoa(cors.AllowMaxAge)

	return func(ctx *atreugo.RequestCtx) error {
		origin := string(ctx.Request.Header.Peek(originHeaderName))

		if !isAllowedOrigin(cors.AllowedOrigins, origin) {
			return ctx.Next()
		}

		ctx.Response.Header.Set(allowOriginHeaderName, origin)

		if cors.AllowCredentials {
			ctx.Response.Header.Set(allowCredentialsHeaderName, "true")
		}

		varyHeader := ctx.Response.Header.Peek(varyHeaderName)
		if len(varyHeader) > 0 {
			varyHeader = append(varyHeader, strHeaderDelim...)
		}

		varyHeader = append(varyHeader, originHeaderName...)
		ctx.Response.Header.SetBytesV(varyHeaderName, varyHeader)

		if len(cors.ExposedHeaders) > 0 {
			ctx.Response.Header.Set(exposeHeadersHeaderName, exposedHeaders)
		}

		method := string(ctx.Method())
		if method != fasthttp.MethodOptions {
			return ctx.Next()
		}

		if len(cors.AllowedHeaders) > 0 {
			ctx.Response.Header.Set(allowHeadersHeaderName, allowedHeaders)
		}

		if len(cors.AllowedMethods) > 0 {
			ctx.Response.Header.Set(allowMethodsHeaderName, allowedMethods)
		}

		if cors.AllowMaxAge > 0 {
			ctx.Response.Header.Set(maxAgeHeaderName, maxAge)
		}

		return ctx.Next()
	}
}

func isAllowedOrigin(allowed []string, origin string) bool {
	for _, v := range allowed {
		if v == origin || v == "*" {
			return true
		}
	}

	return false
}
