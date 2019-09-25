package main

import (
	"github.com/savsgio/atreugo/v8"
	"github.com/savsgio/go-logger"
)

func beforeMiddleware(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed BEFORE view")

	return ctx.Next()
}

func afterMiddleware(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed AFTER view")

	return ctx.Next()
}
