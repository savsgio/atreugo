package atreugo

import (
	"encoding/json"
	"io"

	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
)

func defaultJSONMarshalFunc(w io.Writer, body interface{}) error {
	return json.NewEncoder(w).Encode(body) // nolint:wrapcheck
}

// JSONResponse return response with body in json format.
func (ctx *RequestCtx) JSONResponse(body interface{}, statusCode ...int) error {
	ctx.Response.Header.SetContentType("application/json")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	w := ctx.Response.BodyWriter()

	return ctx.jsonMarshalFunc(w, body) // nolint:wrapcheck
}

// HTTPResponse return response with body in html format.
func (ctx *RequestCtx) HTTPResponse(body string, statusCode ...int) error {
	ctx.Response.Header.SetContentType("text/html; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBodyString(body)

	return nil
}

// HTTPResponseBytes return response with body in html format.
func (ctx *RequestCtx) HTTPResponseBytes(body []byte, statusCode ...int) error {
	ctx.Response.Header.SetContentType("text/html; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBody(body)

	return nil
}

// TextResponse return response with body in text format.
func (ctx *RequestCtx) TextResponse(body string, statusCode ...int) error {
	ctx.Response.Header.SetContentType("text/plain; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBodyString(body)

	return nil
}

// TextResponseBytes return response with body in text format.
func (ctx *RequestCtx) TextResponseBytes(body []byte, statusCode ...int) error {
	ctx.Response.Header.SetContentType("text/plain; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBody(body)

	return nil
}

// RawResponse returns response without encoding the body.
func (ctx *RequestCtx) RawResponse(body string, statusCode ...int) error {
	ctx.Response.Header.SetContentType("application/octet-stream")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBodyString(body)

	return nil
}

// RawResponseBytes returns response without encoding the body.
func (ctx *RequestCtx) RawResponseBytes(body []byte, statusCode ...int) error {
	ctx.Response.Header.SetContentType("application/octet-stream")

	if len(statusCode) > 0 {
		ctx.Response.Header.SetStatusCode(statusCode[0])
	}

	ctx.Response.SetBody(body)

	return nil
}

// FileResponse return a streaming response with file data.
func (ctx *RequestCtx) FileResponse(fileName, filePath, mimeType string) error {
	fasthttp.ServeFile(ctx.RequestCtx, filePath)

	buff := bytebufferpool.Get()
	buff.SetString("attachment; filename=")
	buff.WriteString(fileName) // nolint:errcheck

	ctx.Response.Header.Set("Content-Disposition", buff.String())
	ctx.SetContentType(mimeType)

	bytebufferpool.Put(buff)

	return nil
}

// RedirectResponse redirect request to an especific url.
func (ctx *RequestCtx) RedirectResponse(url string, statusCode int) error {
	ctx.Redirect(url, statusCode)

	return nil
}

// ErrorResponse returns an error response.
func (ctx *RequestCtx) ErrorResponse(err error, statusCode ...int) error {
	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}

	return err
}
