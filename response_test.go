package atreugo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/erikdubbelboer/fasthttp"
)

func Test_newResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
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
				ctx:         new(fasthttp.RequestCtx),
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
				ctx:         new(fasthttp.RequestCtx),
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
			newResponse(tt.args.ctx, tt.args.contentType, tt.args.statusCode...)

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestJSONResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
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
				ctx:        new(fasthttp.RequestCtx),
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
			if err := JSONResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("JSONResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestHTTPResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
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
				ctx:        new(fasthttp.RequestCtx),
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
			if err := HTTPResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("HTTPResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestTextResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
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
				ctx:        new(fasthttp.RequestCtx),
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
			if err := TextResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("TextResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestRawResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
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
				ctx:        new(fasthttp.RequestCtx),
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
			if err := RawResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("RawResponse() error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestFileResponse(t *testing.T) {
	type args struct {
		ctx      *fasthttp.RequestCtx
		fileName string
		filePath string
		mimeType string
	}
	type want struct {
		body        string
		statusCode  int
		contentType string
		getErr      bool
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
				ctx:      new(fasthttp.RequestCtx),
				fileName: "test.pdf",
				filePath: "/tmp/testfile.pdf",
				mimeType: "application/pdf",
			},
			want: want{
				body:        string(testFileContent),
				statusCode:  200,
				contentType: "application/pdf",
				getErr:      false,
			},
		},
		{
			name: "ReadFileError",
			args: args{
				ctx:      new(fasthttp.RequestCtx),
				fileName: "test.pdf",
				filePath: "/blabla/testfile.pdf",
				mimeType: "application/pdf",
			},
			want: want{
				body:        string(testFileContent),
				statusCode:  200,
				contentType: "application/pdf",
				getErr:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioutil.WriteFile(tt.args.filePath, testFileContent, 0644)
			defer os.Remove(tt.args.filePath)

			err := FileResponse(tt.args.ctx, tt.args.fileName, tt.args.filePath, tt.args.mimeType)
			if tt.want.getErr {
				if err == nil {
					t.Error("Error expected")
				}
				// Not check any more
				return
			} else if !tt.want.getErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body: '%v', want: '%v'", responseBody, tt.want.body)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("Header content-type: '%v', want: '%v'", responseContentType, tt.want.contentType)
			}

			wantContentDisposition := fmt.Sprintf("attachment; filename=%s", tt.args.fileName)
			responseContentDisposition := string(tt.args.ctx.Response.Header.Peek("Content-Disposition"))
			if responseContentDisposition != wantContentDisposition {
				t.Errorf("Header content-disposition: '%v', want: '%v'", responseContentDisposition, wantContentDisposition)
			}
		})
	}
}

func TestRedirectResponse(t *testing.T) {
	type args struct {
		ctx        *fasthttp.RequestCtx
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
				ctx:        new(fasthttp.RequestCtx),
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
			if err := RedirectResponse(tt.args.ctx, tt.args.url, tt.args.statusCode); err != nil {
				t.Errorf("RedirectResponse() error: %v", err)
			}

			responseLocation := string(tt.args.ctx.Response.Header.Peek("Location"))
			if responseLocation != tt.want.locationURL {
				t.Errorf("Header content-disposition: '%v', want: '%v'", responseLocation, tt.want.locationURL)
			}

			responseStatusCode := tt.args.ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("status_code: '%v', want: '%v'", responseStatusCode, tt.want.statusCode)
			}
		})
	}
}
