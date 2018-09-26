package atreugo

import (
	"encoding/json"

	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
)

func (ctx *RequestCtx) newResponse(contentType string, statusCode ...int) {
	ctx.SetContentType(contentType)

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	ctx.ResetBody()
}

// JSONResponse return response with body in json format
func (ctx *RequestCtx) JSONResponse(body interface{}, statusCode ...int) error {
	ctx.newResponse("application/json", statusCode...)

	return json.NewEncoder(ctx).Encode(body)
}

// HTTPResponse return response with body in html format
func (ctx *RequestCtx) HTTPResponse(body string, statusCode ...int) error {
	ctx.newResponse("text/html; charset=utf-8", statusCode...)

	_, err := ctx.WriteString(body)
	return err
}

// HTTPResponseBytes return response with body in html format
func (ctx *RequestCtx) HTTPResponseBytes(body []byte, statusCode ...int) error {
	ctx.newResponse("text/html; charset=utf-8", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// TextResponse return response with body in text format
func (ctx *RequestCtx) TextResponse(body string, statusCode ...int) error {
	ctx.newResponse("text/plain; charset=utf-8", statusCode...)

	_, err := ctx.WriteString(body)
	return err
}

// TextResponseBytes return response with body in text format
func (ctx *RequestCtx) TextResponseBytes(body []byte, statusCode ...int) error {
	ctx.newResponse("text/plain; charset=utf-8", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// RawResponse returns response without encoding the body.
func (ctx *RequestCtx) RawResponse(body string, statusCode ...int) error {
	ctx.newResponse("application/octet-stream", statusCode...)

	_, err := ctx.WriteString(body)
	return err
}

// RawResponseBytes returns response without encoding the body.
func (ctx *RequestCtx) RawResponseBytes(body []byte, statusCode ...int) error {
	ctx.newResponse("application/octet-stream", statusCode...)

	_, err := ctx.Write(body)
	return err
}

// FileResponse return a streaming response with file data.
func (ctx *RequestCtx) FileResponse(fileName, filePath, mimeType string) error {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	fasthttp.ServeFile(ctx.RequestCtx, filePath)

	buff.SetString("attachment; filename=")
	buff.WriteString(fileName)

	ctx.Response.Header.Set("Content-Disposition", buff.String())
	ctx.SetContentType(mimeType)

	return nil
}

// RedirectResponse redirect request to an especific url
func (ctx *RequestCtx) RedirectResponse(url string, statusCode int) error {
	ctx.ResetBody()
	ctx.Redirect(url, statusCode)

	return nil
}
