package atreugo

import (
	"context"
	"reflect"
	"testing"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func Test_AcquireRequestCtx(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx) // nolint:ifshort

	if !isEqual(actx.jsonMarshalFunc, defaultJSONMarshalFunc) {
		t.Errorf("jsonMarshalFunc = %p, want %p", actx.jsonMarshalFunc, defaultJSONMarshalFunc)
	}

	if actx.RequestCtx != ctx {
		t.Errorf("RequestCtx = %p, want %p", actx.RequestCtx, ctx)
	}
}

func Test_ReleaseRequestCtx(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	if err := actx.Next(); err != nil {
		t.Fatalf("Error calling next. %+v", err)
	}

	actx.SkipView()

	ReleaseRequestCtx(actx)

	if actx.next {
		t.Errorf("next is not 'false'")
	}

	if actx.skipView {
		t.Errorf("skipView is not 'false'")
	}

	if actx.jsonMarshalFunc == nil {
		t.Errorf("jsonMarshalFunc is nil")
	}

	if actx.RequestCtx != nil {
		t.Errorf("*fasthttp.RequestCtx = %p, want %v", actx.RequestCtx, nil)
	}
}

func TestRequestCtx_RequestID(t *testing.T) {
	value := "123bnj3r2j3rj23"

	ctx := new(RequestCtx)
	ctx.RequestCtx = new(fasthttp.RequestCtx)
	ctx.Request.Header.Set(XRequestIDHeader, value)

	currentValue := string(ctx.RequestID())
	if currentValue != value {
		t.Errorf("ctx.RequestID() = '%s', want '%s'", currentValue, value)
	}
}

func Test_Next(t *testing.T) {
	ctx := AcquireRequestCtx(new(fasthttp.RequestCtx))

	if err := ctx.Next(); err != nil {
		t.Errorf("ctx.Next() unexpected error: %v", err)
	}
}

func Test_SkipView(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	actx.SkipView()

	if !actx.skipView {
		t.Error("ctx.SkipView() is not true")
	}
}

func Test_AttachContext(t *testing.T) {
	type key struct{}

	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	actx.AttachContext(context.WithValue(ctx, key{}, "value"))

	if actx.UserValue(attachedCtxKey) == nil {
		t.Error("ctx.AttachContext() the context is not attached")
	}

	err := catchPanic(func() {
		actx.AttachContext(actx)
	})

	if err == nil {
		t.Error("Panic expected when attachs to itself")
	}
}

func Test_AttachedContext(t *testing.T) {
	type key struct{}

	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)
	otherCtx := context.WithValue(ctx, key{}, "value")

	attachedCtx := actx.AttachedContext()
	if attachedCtx != nil {
		t.Errorf("ctx.AttachedContext() == %p, want %v", attachedCtx, nil)
	}

	actx.AttachContext(otherCtx)

	attachedCtx = actx.AttachedContext()

	if reflect.ValueOf(attachedCtx).Pointer() != reflect.ValueOf(otherCtx).Pointer() {
		t.Errorf("ctx.AttachedContext() == %p, want %p", attachedCtx, otherCtx)
	}
}

func Test_MatchedRoutePath(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	result := actx.MatchedRoutePath()
	if result != nil {
		t.Errorf("ctx.MatchedRoutePathParam() == %s, want %v", result, nil)
	}

	matchedRoutePath := "/user/{id}"

	actx.SetUserValue(router.MatchedRoutePathParam, matchedRoutePath)

	result = actx.MatchedRoutePath()
	if string(result) != matchedRoutePath {
		t.Errorf("ctx.MatchedRoutePathParam() == %s, want %s", result, matchedRoutePath)
	}
}

func Test_Value(t *testing.T) {
	type key struct{}

	structKey := key{}
	stringKey := "stringKey"
	value := "value"

	ctx := new(fasthttp.RequestCtx)
	actx := AcquireRequestCtx(ctx)

	if v := actx.Value("fake"); v != nil {
		t.Errorf("Value() of key '%v' == %s, want %v", "fake", v, nil)
	}

	actx.AttachContext(context.WithValue(actx, structKey, value))
	actx.SetUserValue(stringKey, value)

	if v := actx.Value(structKey); v != value {
		t.Errorf("Value() of key '%v' == %s, want %s", structKey, v, value)
	}

	if v := actx.Value(stringKey); v != value {
		t.Errorf("Value() of key '%s' == %s, want %s", stringKey, v, value)
	}

	if v := actx.Value("fake"); v != nil {
		t.Errorf("Value() key '%s' == %v, want %v", "fake", v, nil)
	}
}
