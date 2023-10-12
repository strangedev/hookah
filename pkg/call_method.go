package pkg

import "reflect"

func callMethod[TImpl any](impl TImpl, method reflect.Method, isIndirect bool, args []any) ReturnValues {
	inputs := make([]reflect.Value, len(args)+1)
	if isIndirect {
		inputs[0] = reflect.ValueOf(&impl)
	} else {
		inputs[0] = reflect.ValueOf(impl)
	}
	for i, arg := range args {
		inputs[i+1] = reflect.ValueOf(arg)
	}

	return method.Func.Call(inputs)
}
