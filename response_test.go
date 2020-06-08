package atreugo

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestJSONResponse(t *testing.T) { //nolint:funlen
	type args struct {
		body       interface{}
		statusCode int
	}

	type want struct {
		body        string
		statusCode  int
		contentType string
		err         bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ValidBody",
			args: args{
				body:       JSON{"test": true},
				statusCode: 200,
			},
			want: want{
				body:        "{\"test\":true}",
				statusCode:  200,
				contentType: "application/json",
				err:         false,
			},
		},
		{
			name: "InvalidBody",
			args: args{
				body:       make(chan int),
				statusCode: 200,
			},
			want: want{
				body:        "",
				statusCode:  200,
				contentType: "application/json",
				err:         true,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := AcquireRequestCtx(ctx)

			err := actx.JSONResponse(tt.args.body, tt.args.statusCode)
			if tt.want.err && (err == nil) {
				t.Errorf("JSONResponse() Expected error")
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

func TestResponses(t *testing.T) { // nolint:funlen
	type args struct {
		fn      func(string, ...int) error
		fnBytes func([]byte, ...int) error
	}

	type want struct {
		body        string
		statusCode  int
		contentType string
	}

	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)
	body := "<h1>Test</h1>"
	statusCode := 403

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "HTTP",
			args: args{
				fn:      actx.HTTPResponse,
				fnBytes: actx.HTTPResponseBytes,
			},
			want: want{
				body:        body,
				statusCode:  statusCode,
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name: "Text",
			args: args{
				fn:      actx.TextResponse,
				fnBytes: actx.TextResponseBytes,
			},
			want: want{
				body:        body,
				statusCode:  statusCode,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Raw",
			args: args{
				fn:      actx.RawResponse,
				fnBytes: actx.RawResponseBytes,
			},
			want: want{
				body:        body,
				statusCode:  statusCode,
				contentType: "application/octet-stream",
			},
		},
	}

	var checkResponse = func(wantBody, wantContentType string, wantStatusCode int) {
		responseBody := string(bytes.TrimSpace(actx.Response.Body()))
		if responseBody != wantBody {
			t.Errorf("body: '%v', want: '%v'", responseBody, wantBody)
		}

		responseStatusCode := actx.Response.StatusCode()
		if responseStatusCode != wantStatusCode {
			t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, wantStatusCode)
		}

		responseContentType := string(actx.Response.Header.ContentType())
		if responseContentType != wantContentType {
			t.Errorf("content-type: '%v', want: '%v'", responseContentType, wantContentType)
		}

		ctx.Request.Reset()
		ctx.Response.Reset()
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.fn(body, statusCode); err != nil {
				t.Errorf("%sResponse() error: %v", tt.name, err)
			}

			checkResponse(tt.want.body, tt.want.contentType, tt.want.statusCode)

			if err := tt.args.fnBytes([]byte(body), statusCode); err != nil {
				t.Errorf("%sResponse() error: %v", tt.name, err)
			}

			checkResponse(tt.want.body, tt.want.contentType, tt.want.statusCode)
		})
	}
}

func TestFileResponse(t *testing.T) { // nolint:funlen
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

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

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			if err := ioutil.WriteFile(tt.args.filePath, testFileContent, 0600); err != nil {
				t.Fatalf("Error writing file %s", tt.args.filePath)
			}
			defer os.Remove(tt.args.filePath)

			ctx := new(fasthttp.RequestCtx)
			actx := AcquireRequestCtx(ctx)

			if err := actx.FileResponse(tt.args.fileName, tt.args.filePath, tt.args.mimeType); err != nil {
				t.Fatalf("Error creating FileResponse for %s", tt.args.fileName)
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

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := AcquireRequestCtx(ctx)

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

func Test_ErrorResponse(t *testing.T) {
	type args struct {
		statusCode []int
	}

	type want struct {
		statusCode int
	}

	err := errors.New("test error")
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "WithStatusCode",
			args: args{
				statusCode: []int{fasthttp.StatusBadRequest},
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "WithOutStatusCode",
			args: args{
				statusCode: make([]int, 0),
			},
			want: want{
				statusCode: fasthttp.StatusInternalServerError,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)
			actx := AcquireRequestCtx(ctx)

			if actx.ErrorResponse(err, tt.args.statusCode...) != err {
				t.Errorf("Unexpected error == %v", err)
			}

			responseStatusCode := actx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}
		})
	}
}

// Benchmarks.
func Benchmark_FileResponse(b *testing.B) {
	cwd, _ := os.Getwd()
	path := cwd + "/LICENSE"

	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		if err := actx.FileResponse("hola", path, "text/plain"); err != nil {
			b.Fatalf("Error calling FileResponse. %+v", err)
		}
	}
}

func Benchmark_JSONResponse(b *testing.B) {
	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	body := JSON{
		"hello":  11,
		"friend": "ascas6d34534rtf3q·$·$&·$&$&&$/&(XCCVasdfasgfds",
		"jsonData": JSON{
			"111": 24.3,
			"asdasdasd23423end3in32im13dfc23fc2 fcec2c": ctx,
		},
	}

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		if err := actx.JSONResponse(body); err != nil {
			b.Fatalf("Error calling JSONResponse. %+v", err)
		}
	}
}
