package atreugo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/erikdubbelboer/fasthttp"
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
func FileResponse(ctx *fasthttp.RequestCtx, fileName, filePath string, mimeType string) error {
	f := atreugoPools.acquireFile()
	defer atreugoPools.putFile(f)

	reader := atreugoPools.acquireBufioReader()
	defer atreugoPools.putBufioReader(reader)

	var err error

	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType(mimeType)

	f, err = os.Open(filePath)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}

	reader = bufio.NewReaderSize(f, int64ToInt(info.Size()))
	ctx.SetBodyStream(reader, int64ToInt(info.Size()))

	return nil
}

// RedirectResponse redirect request to an especific url
func RedirectResponse(ctx *fasthttp.RequestCtx, url string, statusCode int) error {
	ctx.ResetBody()
	ctx.Redirect(url, statusCode)

	return nil
}
