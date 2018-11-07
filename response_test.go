package atreugo

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_newResponse(t *testing.T) {
	type args struct {
		contentType string
		statusCode  []int
	}
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "WithStatusCode",
			args: args{
				contentType: "text/plain",
				statusCode:  []int{301},
			},
			want: want{
				contentType: "text/plain",
				statusCode:  301,
			},
		},
		{
			name: "WithOutStatusCode",
			args: args{
				contentType: "text/plain",
				statusCode:  make([]int, 0),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			actx.newResponse(tt.args.contentType, tt.args.statusCode...)

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestJSONResponse(t *testing.T) {
	type args struct {
		body        interface{}
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       JSON{"test": true},
				statusCode: 200,
			},
			want: want{
				body:        "{\"test\":true}",
				statusCode:  200,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.JSONResponse(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("JSONResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestHTTPResponse(t *testing.T) {
	type args struct {
		body        string
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       "<h1>Test</h1>",
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "text/html; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.HTTPResponse(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("HTTPResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestHTTPResponseBytes(t *testing.T) {
	type args struct {
		body        []byte
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       []byte("<h1>Test</h1>"),
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "text/html; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.HTTPResponseBytes(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("HTTPResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestTextResponse(t *testing.T) {
	type args struct {
		body        string
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       "<h1>Test</h1>",
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.TextResponse(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("TextResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestTextResponseBytes(t *testing.T) {
	type args struct {
		body        []byte
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       []byte("<h1>Test</h1>"),
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.TextResponseBytes(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("TextResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestRawResponse(t *testing.T) {
	type args struct {
		body        string
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       "<h1>Test</h1>",
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "application/octet-stream",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.RawResponse(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("RawResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestRawResponseBytes(t *testing.T) {
	type args struct {
		body        []byte
		statusCode  int
		contentType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				body:       []byte("<h1>Test</h1>"),
				statusCode: 200,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  200,
				contentType: "application/octet-stream",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.RawResponseBytes(tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("RawResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestFileResponse(t *testing.T) {
	type args struct {
		fileName string
		filePath string
		mimeType string
	}
	type want struct {
		body               string
		statusCode         int
		contentType        string
		contentDisposition string
	}

	testFileContent := []byte("Test file content")
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Ok",
			args: args{
				fileName: "test.pdf",
				filePath: "/tmp/testfile.pdf",
				mimeType: "application/pdf",
			},
			want: want{
				body:               string(testFileContent),
				statusCode:         200,
				contentType:        "application/pdf",
				contentDisposition: "attachment; filename=test.pdf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioutil.WriteFile(tt.args.filePath, testFileContent, 0644)
			defer os.Remove(tt.args.filePath)

			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			actx.FileResponse(tt.args.fileName, tt.args.filePath, tt.args.mimeType)

			responseBody := string(bytes.TrimSpace(actx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(actx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("Header content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}

			responseContentDisposition := string(actx.Response.Header.Peek("Content-Disposition"))
			if responseContentDisposition != tt.want.contentDisposition {
				t.Errorf("Header content-disposition: '%v', want: '%v'", responseContentDisposition, tt.want.contentDisposition)
			}
		})
	}
}

func TestRedirectResponse(t *testing.T) {
	type args struct {
		url        string
		statusCode int
	}
	type want struct {
		locationURL string
		statusCode  int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test",
			args: args{
				url:        "http://urltoredirect.es",
				statusCode: 301,
			},
			want: want{
				locationURL: "http://urltoredirect.es/",
				statusCode:  301,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := acquireRequestCtx(ctx)

			if err := actx.RedirectResponse(tt.args.url, tt.args.statusCode); err != nil {
				t.Errorf("RedirectResponse() error: %v", err)
			}

			responseLocation := string(actx.Response.Header.Peek("Location"))
			if responseLocation != tt.want.locationURL {
				t.Errorf("Header content-disposition: '%v', want: '%v'", responseLocation, tt.want.locationURL)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}
		})
	}
}

// Benchmarks
func Benchmark_FileResponse(b *testing.B) {
	cwd, _ := os.Getwd()
	path := cwd + "/LICENSE"

	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		actx.FileResponse("hola", path, "text/plain")
	}
}

func BenchmarkJsoniter_JSONResponse(b *testing.B) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	body := JSON{
		"hello":  11,
		"friend": "ascas6d34534rtf3q·$·$%&·$%&$&&$/%&(XCCVasdfasgfds",
		"jsonData": JSON{
			"111": 24.3,
			"asdasdasd23423end3in32im13dfc23fc2 fcec2c": ctx,
		},
	}

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		actx.JSONResponse(body)
	}
}
