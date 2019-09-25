package main

import (
	"github.com/savsgio/atreugo/v9"
	"github.com/savsgio/go-logger"
)

func beforeFilter(ctx *atreugo.RequestCtx) error {
	logger.Info("Filter executed BEFORE view")

	return ctx.Next()
}

func afterFilter(ctx *atreugo.RequestCtx) error {
	logger.Info("Filter executed AFTER view")

	return ctx.Next()
}
