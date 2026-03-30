// Copyright (c) 2025 Pavlo Moisieienko. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license which can be found in the LICENSE file.

// Package ldi implements lightweight dependency injection
package ldi

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

// New creates new Di
func New() *Di {
	return NewWithParent(nil)
}

// NewWithParent creates new Di with parent
func NewWithParent(parent *Di) *Di {
	return &Di{
		providers: newProviders(),
		parent:    parent,
		resolving: make(map[reflect.Type]struct{}),
	}
}

// Di provides dependency injection
type Di struct {
	providers providers
	parent    *Di
	// Use RWMutex for better read/write performance
	mu sync.RWMutex
	// track in-flight provider resolutions to detect circular dependencies
	resolving map[reflect.Type]struct{}
}

// MustInvoke calls the provided functions if there is error it will panic
func (d *Di) MustInvoke(functions ...any) *Di {
	if err := d.Invoke(functions...); err != nil {
		//revive:disable
		log.Fatal(err)
		//revive:enable
	}
	return d
}

// Invoke calls the provided functions
func (d *Di) Invoke(functions ...any) error {
	for _, function := range functions {
		_, err := d.innerInvoke(function)
		if err != nil {
			d.CleanupResolutionTracking()
			return err
		}
	}
	return nil
}

// MustProvide adds a new provider for the provided value if there is error it will panic
func (d *Di) MustProvide(provide any) *Di {
	if err := d.Provide(provide); err != nil {
		//revive:disable
		log.Fatal(err)
		//revive:enable
	}
	return d
}

// Provide adds a new provider for the provided value
func (d *Di) Provide(anything any) error {
	val := reflect.ValueOf(anything)
	if val.Kind() == reflect.Invalid {
		return fmt.Errorf("can't provide invalid value '%v'", anything)
	}
	if val.Kind() == reflect.Func {
		if val.IsNil() {
			return fmt.Errorf("can't provide function: '%v', type: '%v'", anything, val.Type())
		}
		return d.provideFunction(anything)
	}
	return d.provideValue(val)
}

func (d *Di) provideFunction(function any) error {
	funcType := reflect.TypeOf(function)
	count := 0
	for i := 0; i < funcType.NumOut(); i++ {
		if err := d.provideFunctionValue(function, funcType.Out(i), i); err != nil {
			return fmt.Errorf("failed to add function '%s' as provider: %w", funcType, err)
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("function '%s' must return at least one value", funcType)
	}
	return nil
}
func (d *Di) provideFunctionValue(function any, parameterType reflect.Type, parameterIndex int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	ok, err := d.canAddProvider(parameterType)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return d.providers.addFunction(function, parameterType, parameterIndex)
}

func (d *Di) provideValue(value reflect.Value) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	valueType := value.Type()
	ok, err := d.canAddProvider(valueType)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("couldn't provide value of type '%s'", valueType)
	}
	return d.providers.addValue(value)
}

func (d *Di) innerInvoke(function any) ([]reflect.Value, error) {
	var functionType reflect.Type
	funcValue, ok := function.(reflect.Value)
	if ok && funcValue.IsValid() {
		functionType = funcValue.Type()
	} else {
		functionType = reflect.TypeOf(function)
		funcValue = reflect.ValueOf(function)
	}
	if functionType == nil || functionType.Kind() != reflect.Func {
		return nil, fmt.Errorf("can't invoke not a function '%s'", functionType)
	}

	parameterValues := make([]reflect.Value, 0, functionType.NumIn())
	for i := 0; i < functionType.NumIn(); i++ {
		paramValue, err := d.provideParameterAndCheck(d, functionType.In(i), i)
		if err != nil {
			return nil, fmt.Errorf("function '%s' %w", functionType, err)
		}
		parameterValues = append(parameterValues, paramValue)
	}
	return functionCall(funcValue, parameterValues)
}

func (d *Di) provideParameterAndCheck(di *Di, parameterType reflect.Type, parameterIndex int) (reflect.Value, error) {
	paramValue, err := d.provideParameter(di, parameterType, parameterIndex)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("couldn't provide parameter: %w", err)
	}
	if paramValue.Kind() == reflect.Invalid {
		return paramValue, fmt.Errorf("parameter %d of type '%s' not found", parameterIndex, parameterType)
	}
	return paramValue, nil
}

func (d *Di) provideParameter(di *Di, parameterType reflect.Type, parameterIndex int) (reflect.Value, error) {
	d.mu.RLock()

	if _, resolving := d.resolving[parameterType]; resolving {
		d.mu.RUnlock()
		return reflect.Value{}, fmt.Errorf("circular dependency detected for type '%s'", parameterType)
	}

	prov, ok := d.providers.getProvider(parameterType)
	d.mu.RUnlock()

	if !ok {
		if d.parent != nil {
			return d.parent.provideParameter(di, parameterType, parameterIndex)
		}
		return reflect.Value{}, fmt.Errorf("provider for parameter[%d] of type '%s' not found",
			parameterIndex, parameterType)
	}

	d.mu.Lock()
	// Double-check after acquiring write lock
	if _, resolving := d.resolving[parameterType]; resolving {
		d.mu.Unlock()
		return reflect.Value{}, fmt.Errorf("circular dependency detected for type '%s'", parameterType)
	}
	d.resolving[parameterType] = struct{}{}
	d.mu.Unlock()

	result, err := prov.provide(di)

	// Clean up resolution tracking (always clean up, even on error)
	d.mu.Lock()
	delete(d.resolving, parameterType)
	d.mu.Unlock()

	return result, err
}

func (d *Di) cleanupResolutionTrackingWithParents() {
	for d.parent != nil {
		d.parent.cleanupResolutionTrackingWithParents()
	}
	d.CleanupResolutionTracking()
}

// Clear removes all providers and resets the DI container state
func (d *Di) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.providers = make(providers)
	d.resolving = make(map[reflect.Type]struct{})
}

// CleanupResolutionTracking removes all resolution tracking entries to prevent memory leaks.
// This should be called after error scenarios where tracking might not be cleaned up.
func (d *Di) CleanupResolutionTracking() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear all in-flight resolutions
	d.resolving = make(map[reflect.Type]struct{})
}

func (d *Di) canAddProvider(tp reflect.Type) (bool, error) {
	if isError(tp) {
		return false, nil
	}
	if ok := d.providers.contains(tp); ok {
		return false, fmt.Errorf("provider for type '%s' already exists", tp)
	}
	return true, nil
}
