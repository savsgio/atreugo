package main

import (
	"github.com/savsgio/atreugo/v8"
	"github.com/savsgio/go-logger"
)

func beforeFilter(ctx *atreugo.RequestCtx) (int, error) {
	logger.Info("Filter executed BEFORE view")

	return 200, nil
}

func afterFilter(ctx *atreugo.RequestCtx) (int, error) {
	logger.Info("Filter executed AFTER view")

	beforeFilter(ctx)

	return 200, nil
}
