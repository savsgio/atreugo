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
		options := middlewares.CorsOptions{
			AllowedOrigins:   []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
			AllowedHeaders:   []string{"Content-Type", "content-type"},
			AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
			AllowCredentials: true,
			AllowMaxAge:      5600,
		}
		if err := middlewares.NewCorsMiddleware(options); err != nil {
			logger.Error("server.UseBefore executed NewCorsMiddleware")
		}

		return ctx.Next()
	})

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
