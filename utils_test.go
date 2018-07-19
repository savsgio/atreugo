package atreugo

import (
	"bufio"
	"os"
	"reflect"
	"testing"
)

func Test_pools_acquireFile(t *testing.T) {
	tests := []struct {
		name string
		p    *pools
		want *os.File
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.acquireFile(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pools.acquireFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pools_acquireBufioReader(t *testing.T) {
	tests := []struct {
		name string
		p    *pools
		want *bufio.Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.acquireBufioReader(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pools.acquireBufioReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pools_putFile(t *testing.T) {
	type args struct {
		f *os.File
	}
	tests := []struct {
		name string
		p    *pools
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.putFile(tt.args.f)
		})
	}
}

func Test_pools_putBufioReader(t *testing.T) {
	type args struct {
		br *bufio.Reader
	}
	tests := []struct {
		name string
		p    *pools
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.putBufioReader(tt.args.br)
		})
	}
}

func Test_panicOnError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panicOnError(tt.args.err)
		})
	}
}

func Test_b2s(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b2s(tt.args.b); got != tt.want {
				t.Errorf("b2s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_int64ToInt(t *testing.T) {
	type args struct {
		i int64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int64ToInt(tt.args.i); got != tt.want {
				t.Errorf("int64ToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_indexOf(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indexOf(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("indexOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_include(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := include(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("include() = %v, want %v", got, tt.want)
			}
		})
	}
}
