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
	_, _, err := getMethod(h.impl, methodName)
	if err != nil {
		return err
	}

	h.returnHooks[methodName] = hook

	return nil
}

func (h *Hookah[TImpl]) RunMethodWithReturnHooks(methodName string, args ...any) ReturnValues {
	method, isIndirect, err := getMethod(h.impl, methodName)
	if err != nil {
		panic(err)
	}

	originalReturnValues := callMethod(h.impl, method, isIndirect, args)

	if hook, ok := h.returnHooks[methodName]; ok {
		return hook(originalReturnValues)
	}

	return originalReturnValues
}
