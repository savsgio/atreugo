package middlewares

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/valyala/fasthttp"
	"strconv"

	"strings"
)

// Create new properties for CORS.
type CorsOptions struct {
	// given origin or origins can be shared with response
	AllowedOrigins []string

	// which headers can be used during the actual request
	AllowedHeaders []string

	// seconds for caching the preflight result
	AllowMaxAge int

	// specify the method or methods allowed to resource
	AllowedMethods []string

	// which header names can be exposed as part of the response
	ExposedHeaders []string

	// whether to expose the response to frontend code
	AllowCredentials bool
}

type CorsHandler struct {
	allowedOrigins []string
	allowedHeaders []string
	allowedMethods []string
	exposedHeaders []string
	maxAge         int
}

var defaultOptions = CorsOptions{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST"},
	AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
}

func NewCorsMiddleware(options CorsOptions) atreugo.Middleware {
	cors := CorsHandler{
		allowedOrigins: options.AllowedOrigins,
		allowedHeaders: options.AllowedHeaders,
		allowedMethods: options.AllowedMethods,
		exposedHeaders: options.ExposedHeaders,
	}

	for _, v := range options.AllowedOrigins {
		if v == "*" {
			cors.allowedOrigins = defaultOptions.AllowedOrigins
			break
		}
	}

	if len(cors.allowedMethods) == 0 {
		cors.allowedMethods = defaultOptions.AllowedMethods
	}

	if len(cors.allowedMethods) == 0 {
		cors.allowedMethods = defaultOptions.AllowedMethods
	}

	return cors.middleware
}

func (c *CorsHandler) middleware(ctx *atreugo.RequestCtx) error {
	return c.handlePreflight(ctx)
}

func (c *CorsHandler) handlePreflight(ctx *atreugo.RequestCtx) error {
	originHeader := string(ctx.Request.Header.Peek("Origin"))

	if !c.isAllowedOrigin(originHeader) {
		return ctx.Next()
	}

	ctx.Response.Header.Set("Access-Control-Allow-Origin", originHeader)

	varyHeader := ctx.Response.Header.Peek("Vary")
	if len(varyHeader) > 0 {
		varyHeader = append(varyHeader, ", "...)
	}

	varyHeader = append(varyHeader, "Origin"...)

	ctx.Response.Header.SetBytesV("Vary", varyHeader)

	method := string(ctx.Request.Header.Method())
	if method != fasthttp.MethodOptions {
		return ctx.Next()
	}

	if len(c.allowedMethods) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(c.allowedMethods, ", "))
	}

	if len(c.allowedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(c.allowedHeaders, ", "))
	}

	if c.maxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
	}

	return ctx.Next()
}

func (c *CorsHandler) isAllowedOrigin(originHeader string) bool {
	for _, val := range c.allowedOrigins {
		if val == originHeader {
			return true
		}
	}

	return false
}
