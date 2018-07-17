package atreugo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/erikdubbelboer/fasthttp"
)

// JSON is a map whose key is a string and whose value an interface
type JSON map[string]interface{}

// JSONResponse return response with body in json format
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

// HTTPResponse return response with body in html format
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

// TextResponse return response with body in text format
func TextResponse(ctx *fasthttp.RequestCtx, response []byte, statusCode ...int) error {
	ctx.SetContentType("text/plain; charset=utf-8")

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	_, err := ctx.Write(response)
	return err
}

// RawResponse returns response without encoding the body.
func RawResponse(ctx *fasthttp.RequestCtx, response []byte, statusCode ...int) error {
	ctx.SetContentType("application/octet-stream")

	if len(statusCode) > 0 {
		ctx.SetStatusCode(statusCode[0])
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	_, err := ctx.Write(response)
	return err
}

// FileResponse return a streaming response with file data.
func FileResponse(ctx *fasthttp.RequestCtx, fileName, filePath string, mimeType string) error {
	f := atreugoPools.getFile()
	defer atreugoPools.putFile(f)

	reader := atreugoPools.getBufioReader()
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
