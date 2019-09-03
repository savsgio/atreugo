package main

import (
	"github.com/savsgio/atreugo/v8"
	"github.com/savsgio/go-logger"
)

func beforeMiddleware(ctx *atreugo.RequestCtx) (int, error) {
	logger.Info("Middleware executed BEFORE view")

	return 0, nil
}

func afterMiddleware(ctx *atreugo.RequestCtx) (int, error) {
	logger.Info("Middleware executed AFTER view")

	return 0, nil
}
