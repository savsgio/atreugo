package middlewares

import (
	"testing"

	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
)

func TestRequestIDMiddleware(t *testing.T) { //nolint:funlen
	type args struct {
		predefinedRequestID []byte
	}

	type want struct {
		newValue bool
		err      bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Create",
			args: args{
				predefinedRequestID: nil,
			},
			want: want{
				newValue: true,
				err:      false,
			},
		},
		{
			name: "Predefined",
			args: args{
				predefinedRequestID: []byte("1342nwjdviwer3c2e32e"),
			},
			want: want{
				newValue: false,
				err:      false,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			ctx := new(atreugo.RequestCtx)
			ctx.RequestCtx = new(fasthttp.RequestCtx)

			if tt.args.predefinedRequestID != nil {
				ctx.Request.Header.SetBytesV(atreugo.XRequestIDHeader, tt.args.predefinedRequestID)
			}

			err := RequestIDMiddleware(ctx)
			if (err != nil) != tt.want.err {
				t.Errorf("RequestIDMiddleware() unexpected error: %v", err)
				return
			}

			value := ctx.Request.Header.Peek(atreugo.XRequestIDHeader)

			if tt.want.newValue && (string(tt.args.predefinedRequestID) == string(value)) {
				t.Error("RequestIDMiddleware() not set a new value")
			} else if !tt.want.newValue && (string(tt.args.predefinedRequestID) != string(value)) {
				t.Error("RequestIDMiddleware() changed value unexpectly")
			}
		})
	}
}
