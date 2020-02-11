package main

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/go-logger"
)

func beforeGlobal(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed BEFORE GLOBAL")

	return ctx.Next()
}

func afterGlobal(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed AFTER GLOBAL")

	return ctx.Next()
}

func beforeView(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed BEFORE VIEW")

	return ctx.Next()
}

func afterView(ctx *atreugo.RequestCtx) error {
	logger.Info("Middleware executed AFTER VIEW")

	return ctx.Next()
}
