package main

import "github.com/savsgio/atreugo/v10"

// CORS Setting
var (
	corsAllowHeaders     = "Content-Type, authorization"
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

func corsMiddleware(ctx *atreugo.RequestCtx) error {
	// Avoid CORS middleware when you need
	path := string(ctx.Path())
	ignoredURL := map[string]bool{
		"/no-cors": true,
	}
	if ignoredURL[path] {
		return ctx.Next()
	}

	// Dynamic value
	//corsAllowOrigin := string(ctx.URI().Scheme()) + "://" + string(ctx.Host())

	// Mandatory header
	ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)

	// Optional headers
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
	ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)

	return ctx.Next()
}
