package atreugo

import (
	"encoding/json"

	"github.com/erikdubbelboer/fasthttp"
)

// JSON is a map whose key is a string and whose value an interface
type JSON map[string]interface{}

// JSONResponse return http response with content-type application/json
func JSONResponse(ctx *fasthttp.RequestCtx, response interface{}, statusCode ...int) error {
	ctx.SetContentType("application/json")

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	ctx.ResetBody()
	return json.NewEncoder(ctx).Encode(response)
}

// HTTPResponse return http response with content-type text/html; charset=utf-8
func HTTPResponse(ctx *fasthttp.RequestCtx, response []byte, statusCode ...int) error {
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

// RedirectResponse redirect request to an especific url
func RedirectResponse(ctx *fasthttp.RequestCtx, url string, statusCode int) error {
	ctx.ResetBody()
	ctx.Redirect(url, statusCode)

	return nil
}
