package middlewares

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"

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
	allowedHeadersAll bool
	allowedOrigins    []string
	allowedHeaders    []string
	allowedMethods    []string
	exposedHeaders    []string
}

var defaultOptions = CorsOptions{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST"},
	AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
}

func DefaultCors() *CorsHandler {
	return NewCorsMiddleware(defaultOptions)
}

func NewCorsMiddleware(options CorsOptions) *CorsHandler {
	cors := CorsHandler{
		allowedOrigins: options.AllowedOrigins,
		allowedHeaders: options.AllowedHeaders,
		allowedMethods: options.AllowedMethods,
		exposedHeaders: options.ExposedHeaders,
	}

	if len(cors.allowedOrigins) == 0 {
		cors.allowedOrigins = defaultOptions.AllowedOrigins
	} else {
		for _, v := range options.AllowedOrigins {
			if v == "*" {
				cors.allowedOrigins = defaultOptions.AllowedOrigins
				break
			}
		}
	}

	if len(cors.allowedHeaders) == 0 {
		cors.allowedHeaders = defaultOptions.AllowedHeaders
		cors.allowedHeadersAll = true
	} else {
		for _, v := range options.AllowedHeaders {
			if v == "*" {
				cors.allowedHeadersAll = true
				break
			}
		}
	}

	if len(cors.allowedMethods) == 0 {
		cors.allowedMethods = defaultOptions.AllowedMethods
	}

	return &cors
}

func (c *CorsHandler) CorsMiddleware(ctx *atreugo.RequestCtx) error {
	err := c.handlePreflight(ctx)
	if err != nil {
		logger.Error(err)
	}

	return ctx.Next()
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
