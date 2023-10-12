package pkg

import (
	"errors"
	"reflect"
)

type Hookah[TImpl any] struct {
	impl        TImpl
	returnHooks map[string]ReturnHook
}

type ReturnHook func(returnValues ReturnValues) (updatedReturnValues ReturnValues)

type ReturnValues []reflect.Value

func NewHookah[TImpl any](impl TImpl) *Hookah[TImpl] {
	return &Hookah[TImpl]{
		impl:        impl,
		returnHooks: make(map[string]ReturnHook),
	}
}

var (
	ErrMethodNotFound = errors.New("method not found")
)

func (h *Hookah[TImpl]) AddReturnHook(methodName string, hook ReturnHook) error {
	self := reflect.TypeOf(h.impl)
	_, ok := self.MethodByName(methodName)
	if !ok {
		selfPtr := reflect.TypeOf(&h.impl)
		_, ok := selfPtr.MethodByName(methodName)
		if !ok {
			return ErrMethodNotFound
		}
	}

	h.returnHooks[methodName] = hook

	return nil
}

func (h *Hookah[TImpl]) RunMethodWithReturnHooks(methodName string, args ...any) ReturnValues {
	var method reflect.Method
	var ok bool
	var isIndirect bool
	method, ok = reflect.TypeOf(h.impl).MethodByName(methodName)
	if !ok {
		method, ok = reflect.TypeOf(&h.impl).MethodByName(methodName)
		if !ok {
			panic(ErrMethodNotFound)
		}
		isIndirect = true
	}

	inputs := make([]reflect.Value, len(args)+1)
	if isIndirect {
		inputs[0] = reflect.ValueOf(&h.impl)
	} else {
		inputs[0] = reflect.ValueOf(h.impl)
	}
	for i, arg := range args {
		inputs[i+1] = reflect.ValueOf(arg)
	}

	originalReturnValues := method.Func.Call(inputs)

	if hook, ok := h.returnHooks[methodName]; ok {
		return hook(originalReturnValues)
	}

	return originalReturnValues
}
