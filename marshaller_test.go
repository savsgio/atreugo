package atreugo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type CustomJSONMarshaller struct{}

func (m *CustomJSONMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func TestCustomizeJsonMarshaller(t *testing.T) {
	type args struct {
		body           interface{}
		statusCode     int
		jsonMarshaller JSONMarshaller
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
			name: "CustomizeJSONMarshaller",
			args: args{
				body:           JSON{"test": true},
				statusCode:     200,
				jsonMarshaller: &CustomJSONMarshaller{},
			},
			want: want{
				body:       "{\"test\":true}",
				statusCode: 200,
			},
		},
		{
			name: "DefaultJSONMarshaller",
			args: args{
				body:       JSON{"test": true},
				statusCode: 200,
			},
			want: want{
				body:       "{\"test\":true}",
				statusCode: 200,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Addr:           "127.0.0.1:8000",
				JSONMarshaller: tt.args.jsonMarshaller,
			}
			server := New(config)

			server.GET("/json", func(ctx *RequestCtx) error {
				return ctx.JSONResponse(tt.args.body, tt.args.statusCode)
			})
			go func() {
				time.Sleep(2 * time.Second)
				// make http get request to the server
				response, err := http.Get(fmt.Sprintf("http://%s/json", config.Addr))
				if err != nil {
					t.Errorf("Error making http get request: %+v", err)
				}
				//marshal response to json
				resBody, err := io.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("Could not read response body: %s\n", err)
					os.Exit(1)
				}
				if string(resBody) != tt.want.body {
					t.Errorf("Expected response body to be %s, got %s", tt.want.body, string(resBody))
				}
				if response.StatusCode != tt.want.statusCode {
					t.Errorf("Expected response status code to be %d, got %d", tt.want.statusCode, response.StatusCode)
				}
				if err := server.engine.Shutdown(); err != nil {
					t.Errorf("Error shutting down the server: %+v", err)
				}
			}()
			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}
		})
	}
}
