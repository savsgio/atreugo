package middlewares

import (
	"github.com/savsgio/atreugo/v10"

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
	options CorsOptions
}

func NewCorsMiddleware(options CorsOptions) atreugo.Middleware {
	cors := CorsHandler{
		options: options,
	}

	return cors.middleware
}

func (c *CorsHandler) middleware(ctx *atreugo.RequestCtx) error {
	if err := c.handlePreflight(ctx); err != nil {
		return err
	}

	return ctx.Next()
}

func (c *CorsHandler) handlePreflight(ctx *atreugo.RequestCtx) error {
	originHeader := string(ctx.Request.Header.Peek("Origin"))

	if !c.isAllowedOrigin(originHeader) {
		return ctx.Next()
	}

	// Mandatory header
	ctx.Response.Header.Set("Access-Control-Allow-Origin", originHeader)

	if c.options.AllowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	varyHeader := ctx.Response.Header.Peek("Vary")
	if len(varyHeader) > 0 {
		varyHeader = append(varyHeader, ", "...)
	}

	varyHeader = append(varyHeader, "Origin"...)

	ctx.Response.Header.SetBytesV("Vary", varyHeader)

	if len(c.options.AllowedMethods) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(c.options.AllowedMethods, ", "))
	}

	if len(c.options.AllowedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(c.options.AllowedHeaders, ", "))
	}

	if c.options.AllowMaxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", strconv.Itoa(c.options.AllowMaxAge))
	}

	if len(c.options.ExposedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(c.options.ExposedHeaders, ", "))
	}

	return ctx.Next()
}

func (c *CorsHandler) isAllowedOrigin(originHeader string) bool {
	for _, val := range c.options.AllowedOrigins {
		if val == originHeader || val == "*" {
			return true
		}
	}

	return false
}
