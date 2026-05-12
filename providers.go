// Copyright (c) 2025 Pavlo Moisieienko. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license which can be found in the LICENSE file.

package ldi

import (
	"fmt"
	"reflect"
)

var errTp = reflect.TypeOf((*error)(nil)).Elem()

type providerType uint8

const (
	valueProviderType providerType = iota
	functionProviderType
)

type provider struct {
	providerType   providerType
	value          reflect.Value
	function       any
	parameterIndex int
	providers      providers
}

type providers map[reflect.Type]*provider

func (ps providers) contains(tp reflect.Type) bool {
	_, ok := ps[tp]
	return ok
}

func (ps providers) getProvider(tp reflect.Type) (*provider, bool) {
	p, ok := ps[tp]
	return p, ok
}

func (ps providers) addFunction(function any, parameterType reflect.Type, parameterIndex int) error {
	return ps.addProvider(parameterType, newFunctionProvider(function, parameterIndex, ps))
}

func (ps providers) addValue(value reflect.Value) error {
	return ps.addProvider(value.Type(), newValueProvider(value))
}

func (ps providers) addProvider(tp reflect.Type, prov *provider) error {
	if ps.contains(tp) {
		return fmt.Errorf("provider for type '%s' already exists", tp)
	}
	ps[tp] = prov
	return nil
}

func (ps providers) setFunctionProvidersValues(values []reflect.Value) bool {
	result := false
	for i := 0; i < len(values); i++ {
		if pr, ok := ps.getProvider(values[i].Type()); ok {
			if pr.providerType == functionProviderType {
				result = true
				pr.value = values[i]
			}
		}
	}
	return result
}

func (ps providers) Len() int {
	return len(ps)
}

func newProviders() providers {
	// initial value of the number of providers,
	// there is no point in creating a list of providers without the providers themselves
	const initialValue = 1
	return make(providers, initialValue)
}

func (p *provider) provide(di *Di) (reflect.Value, error) {
	switch p.providerType {
	case valueProviderType:
		return p.value, nil
	case functionProviderType:
		if p.value.IsValid() {
			return p.value, nil
		}

		values, err := di.innerInvoke(p.function)
		if err != nil {
			return reflect.Value{}, err
		}

		// set function result values to providers to prevent the function from being called again
		if !p.providers.setFunctionProvidersValues(values) {
			return reflect.Value{}, fmt.Errorf("values of function '%s' did not set", p.function)
		}

		return p.value, nil
	default:
		return reflect.Value{}, fmt.Errorf("unknown provider type: %d", p.providerType)
	}
}

func newFunctionProvider(function any, parameterIndex int, providers providers) *provider {
	return &provider{
		providerType:   functionProviderType,
		function:       function,
		parameterIndex: parameterIndex,
		providers:      providers,
	}
}

func newValueProvider(value reflect.Value) *provider {
	return &provider{
		providerType: valueProviderType,
		value:        value,
	}
}

func isError(tp reflect.Type) bool {
	return tp.Implements(errTp)
}

func functionCall(fnc reflect.Value, args []reflect.Value) ([]reflect.Value, error) {
	results := fnc.Call(args)
	for _, r := range results {
		if isError(r.Type()) {
			if err, ok := r.Interface().(error); ok {
				// if an error occurs when calling a function, we are not interested in other results
				return nil, err
			}
		}
	}
	return results, nil
}
