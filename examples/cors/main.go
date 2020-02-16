package main

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/atreugo/v10/middlewares"
	"github.com/savsgio/go-logger"
)

func init() { //nolint:gochecknoinits
	logger.SetLevel(logger.DEBUG)
}

func main() {
	config := atreugo.Config{
		Addr:     "0.0.0.0:8001",
		LogLevel: "debug",
	}
	server := atreugo.New(config)

	server.UseBefore(func(ctx *atreugo.RequestCtx) error {
		withCors := middlewares.NewCorsMiddleware(middlewares.CorsOptions{
			// if you leave allowedOrigins empty then atreugo will treat it as "*"
			AllowedOrigins: []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
			// if you leave allowedHeaders empty then atreugo will accept any non-simple headers
			AllowedHeaders: []string{"Content-Type", "content-type"},
			// if you leave this empty, only simple method will be accepted
			AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
			AllowCredentials: true,
			AllowMaxAge:      5600,
		})
		err := withCors.CorsMiddleware(ctx)
		if err != nil {
			logger.Error(err)
		}

		return ctx.Next()
	})

	// Use CORS with default options
	//server.UseBefore(func(ctx *atreugo.RequestCtx) error {
	//	withCors := middlewares.DefaultCors()
	//	err := withCors.CorsMiddleware(ctx)
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//
	//	return ctx.Next()
	//})

	server.GET("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Method": "GET"})
	})

	server.POST("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Method": "POST"})
	})

	server.PUT("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Method": "PUT"})
	})

	server.DELETE("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Method": "DELETE"})
	})

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
