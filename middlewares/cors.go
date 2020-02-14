package middlewares

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/go-logger"

	"strconv"
	"strings"
)

type CorsOptions struct {
	AllowedOrigins   []string
	AllowedHeaders   []string
	AllowMaxAge      int
	AllowedMethods   []string
	ExposedHeaders   []string
	AllowCredentials bool
}

type CorsHandler struct {
	maxAge            int
	allowCredentials  bool
	allowedOriginsAll bool
	allowedHeadersAll bool
	allowedOrigins    []string
	allowedHeaders    []string
	allowedMethods    []string
	exposedHeaders    []string
}

var defaultOptions = &CorsOptions{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST"},
	AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
}

func DefaultCors() *CorsHandler {
	return NewCorsMiddleware(*defaultOptions)
}

func NewCorsMiddleware(options CorsOptions) *CorsHandler {
	cors := &CorsHandler{
		allowedOrigins:   options.AllowedOrigins,
		allowedHeaders:   options.AllowedHeaders,
		allowCredentials: options.AllowCredentials,
		allowedMethods:   options.AllowedMethods,
		exposedHeaders:   options.ExposedHeaders,
		maxAge:           options.AllowMaxAge,
	}

	if len(cors.allowedOrigins) == 0 {
		cors.allowedOrigins = defaultOptions.AllowedOrigins
		cors.allowedOriginsAll = true
	} else {
		for _, v := range options.AllowedOrigins {
			if v == "*" {
				cors.allowedOrigins = defaultOptions.AllowedOrigins
				cors.allowedOriginsAll = true
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

	return cors
}

func (c *CorsHandler) CorsMiddleware(ctx *atreugo.RequestCtx) {
	c.handlePreflight(ctx)
}

func (c *CorsHandler) handlePreflight(ctx *atreugo.RequestCtx) {
	originHeader := string(ctx.Request.Header.Peek("Origin"))

	if len(originHeader) == 0 || !c.isAllowedOrigin(originHeader) {
		logger.Debug("Origin ", originHeader, " is not in", c.allowedOrigins)
		return
	}

	method := string(ctx.Request.Header.Method())
	if !c.isAllowedMethod(method) {
		logger.Debug("Method ", method, " is not in", c.allowedMethods)
		return
	}

	var headers []string
	if len(ctx.Request.Header.Peek("Access-Control-Request-Headers")) > 0 {
		headers = strings.Split(string(ctx.Request.Header.Peek("Access-Control-Request-Headers")), ",")
	}

	if !c.areHeadersAllowed(headers) {
		logger.Debug("Headers ", headers, " is not in", c.allowedHeaders)
		return
	}

	ctx.Response.Header.Set("Access-Control-Allow-Origin", originHeader)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", method)

	if len(headers) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
	}

	if c.allowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if c.maxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
	}
}

func (c *CorsHandler) isAllowedOrigin(originHeader string) bool {
	if c.allowedOriginsAll {
		return true
	}

	for _, val := range c.allowedOrigins {
		if val == originHeader {
			return true
		}
	}

	return false
}

func (c *CorsHandler) isAllowedMethod(methodHeader string) bool {
	if len(c.allowedMethods) == 0 {
		return false
	}

	if methodHeader == "OPTIONS" {
		return true
	}

	for _, m := range c.allowedMethods {
		if m == methodHeader {
			return true
		}
	}

	return false
}

func (c *CorsHandler) areHeadersAllowed(headers []string) bool {
	if c.allowedHeadersAll || len(headers) == 0 {
		return true
	}

	for _, header := range headers {
		found := false

		for _, h := range c.allowedHeaders {
			if h == header {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}
