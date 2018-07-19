package atreugo

import (
	"bytes"
	"testing"

	"github.com/erikdubbelboer/fasthttp"
)

func Test_newResponse(t *testing.T) {
	type args struct {
		ctx         *fasthttp.RequestCtx
		contentType string
		statusCode  int
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
				statusCode:  301,
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
			},
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newResponse(tt.args.ctx, tt.args.contentType, tt.args.statusCode)

			if tt.args.ctx.Response.StatusCode() != tt.want.statusCode {
				t.Errorf("status_code = %v, want %v", tt.args.statusCode, tt.want.statusCode)
			}

			if string(tt.args.ctx.Response.Header.ContentType()) != tt.want.contentType {
				t.Errorf("content-type = %v, want %v", tt.args.statusCode, tt.want.statusCode)
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
		body        interface{}
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
				statusCode: 1,
			},
			want: want{
				body:        "{\"test\":true}",
				statusCode:  1,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JSONResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("JSONResponse() error = %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body = %v, want %v", responseBody, tt.want.body)
			}

			if tt.args.ctx.Response.StatusCode() != tt.want.statusCode {
				t.Errorf("status_code = %v, want %v", tt.args.statusCode, tt.want.statusCode)
			}

			if string(tt.args.ctx.Response.Header.ContentType()) != tt.want.contentType {
				t.Errorf("content-type = %v, want %v", tt.args.contentType, tt.want.contentType)
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
		body        interface{}
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
				statusCode: 1,
			},
			want: want{
				body:        "<h1>Test</h1>",
				statusCode:  1,
				contentType: "text/html; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HTTPResponse(tt.args.ctx, tt.args.body, tt.args.statusCode); err != nil {
				t.Errorf("HTTPResponse() error = %v", err)
			}

			responseBody := string(bytes.TrimSpace(tt.args.ctx.Response.Body()))
			if responseBody != tt.want.body {
				t.Errorf("body = %v, want %v", responseBody, tt.want.body)
			}

			if tt.args.ctx.Response.StatusCode() != tt.want.statusCode {
				t.Errorf("status_code = %v, want %v", tt.args.statusCode, tt.want.statusCode)
			}

			responseContentType := string(tt.args.ctx.Response.Header.ContentType())
			if responseContentType != tt.want.contentType {
				t.Errorf("content-type = %v, want %v", responseContentType, tt.want.contentType)
			}
		})
	}
}

func TestTextResponse(t *testing.T) {
	type args struct {
		ctx        *fasthttp.RequestCtx
		body       []byte
		statusCode []int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TextResponse(tt.args.ctx, tt.args.body, tt.args.statusCode...); (err != nil) != tt.wantErr {
				t.Errorf("TextResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRawResponse(t *testing.T) {
	type args struct {
		ctx        *fasthttp.RequestCtx
		body       []byte
		statusCode []int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RawResponse(tt.args.ctx, tt.args.body, tt.args.statusCode...); (err != nil) != tt.wantErr {
				t.Errorf("RawResponse() error = %v, wantErr %v", err, tt.wantErr)
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
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FileResponse(tt.args.ctx, tt.args.fileName, tt.args.filePath, tt.args.mimeType); (err != nil) != tt.wantErr {
				t.Errorf("FileResponse() error = %v, wantErr %v", err, tt.wantErr)
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
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RedirectResponse(tt.args.ctx, tt.args.url, tt.args.statusCode); (err != nil) != tt.wantErr {
				t.Errorf("RedirectResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
