package atreugo

import (
	"net"
	"reflect"
	"testing"

	"github.com/erikdubbelboer/fasthttp"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name string
		args args
		want *Atreugo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtreugo_handler(t *testing.T) {
	type args struct {
		viewFn View
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
		want fasthttp.RequestHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.handler(tt.args.viewFn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Atreugo.handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtreugo_getListener(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
		want net.Listener
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.getListener(tt.args.addr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Atreugo.getListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtreugo_serve(t *testing.T) {
	type args struct {
		ln       net.Listener
		protocol string
		addr     string
	}
	tests := []struct {
		name    string
		s       *Atreugo
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.serve(tt.args.ln, tt.args.protocol, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAtreugo_serveGracefully(t *testing.T) {
	type args struct {
		ln       net.Listener
		protocol string
		addr     string
	}
	tests := []struct {
		name    string
		s       *Atreugo
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.serveGracefully(tt.args.ln, tt.args.protocol, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.serveGracefully() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAtreugo_Static(t *testing.T) {
	type args struct {
		rootStaticDirPath string
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Static(tt.args.rootStaticDirPath)
		})
	}
}

func TestAtreugo_Path(t *testing.T) {
	type args struct {
		httpMethod string
		url        string
		viewFn     View
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Path(tt.args.httpMethod, tt.args.url, tt.args.viewFn)
		})
	}
}

func TestAtreugo_UseMiddleware(t *testing.T) {
	type args struct {
		fns []Middleware
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.UseMiddleware(tt.args.fns...)
		})
	}
}

func TestAtreugo_ListenAndServe(t *testing.T) {
	tests := []struct {
		name    string
		s       *Atreugo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.ListenAndServe(); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
