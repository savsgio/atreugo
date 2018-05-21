package atreugo

import (
	"fmt"
	"reflect"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func callFuncByName(class interface{}, funcName string, params ...interface{}) []reflect.Value {
	fn := reflect.ValueOf(class).MethodByName(funcName)

	if !fn.IsValid() {
		panic(fmt.Errorf("Method not found \"%s\"", funcName))
	}

	args := make([]reflect.Value, len(params))
	for i, param := range params {
		args[i] = reflect.ValueOf(param)
	}

	return fn.Call(args)
}
