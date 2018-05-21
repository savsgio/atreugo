package atreugo

import (
	"encoding/json"

	"github.com/erikdubbelboer/fasthttp"
)

type Json map[string]interface{}

func JsonResponse(ctx *fasthttp.RequestCtx, response interface{}, statusCode ...int) error {
	ctx.SetContentType("application/json")

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	ctx.ResetBody()
	return json.NewEncoder(ctx).Encode(response)
}

func HttpResponse(ctx *fasthttp.RequestCtx, response []byte, statusCode ...int) error {
	ctx.SetContentType("text/html; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	ctx.ResetBody()

	_, err := ctx.Write(response)
	return err
}
