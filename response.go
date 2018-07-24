package atreugo

import (
	"encoding/json"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/valyala/bytebufferpool"
)

func newResponse(ctx *fasthttp.RequestCtx, contentType string, statusCode ...int) {
	ctx.SetContentType(contentType)

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	ctx.ResetBody()
}

// JSONResponse return response with body in json format
func JSONResponse(ctx *fasthttp.RequestCtx, body interface{}, statusCode ...int) error {
	newResponse(ctx, "application/json", statusCode...)

	return json.NewEncoder(ctx).Encode(body)
}

// HTTPResponse return response with body in html format
func HTTPResponse(ctx *fasthttp.RequestCtx, body []byte, statusCode ...int) error {
	newResponse(ctx, "text/html; charset=utf-8", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// TextResponse return response with body in text format
func TextResponse(ctx *fasthttp.RequestCtx, body []byte, statusCode ...int) error {
	newResponse(ctx, "text/plain; charset=utf-8", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// RawResponse returns response without encoding the body.
func RawResponse(ctx *fasthttp.RequestCtx, body []byte, statusCode ...int) error {
	newResponse(ctx, "application/octet-stream", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// FileResponse return a streaming response with file data.
func FileResponse(ctx *fasthttp.RequestCtx, fileName, filePath, mimeType string) error {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	fasthttp.ServeFile(ctx, filePath)

	buff.SetString("attachment; filename=")
	buff.WriteString(fileName)

	ctx.Response.Header.Set("Content-Disposition", buff.String())
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType(mimeType)

	return nil
}

// RedirectResponse redirect request to an especific url
func RedirectResponse(ctx *fasthttp.RequestCtx, url string, statusCode int) error {
	ctx.ResetBody()
	ctx.Redirect(url, statusCode)

	return nil
}
