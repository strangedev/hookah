package pkg

import "reflect"

func getMethod[TImpl any](impl TImpl, methodName string) (method reflect.Method, isIndirect bool, err error) {
	var found bool

	method, found = reflect.TypeOf(impl).MethodByName(methodName)
	if found {
		return method, false, nil
	}

	method, found = reflect.TypeOf(&impl).MethodByName(methodName)
	if found {
		return method, true, nil
	}

	return method, isIndirect, ErrMethodNotFound
}
