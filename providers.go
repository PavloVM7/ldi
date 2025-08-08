// Copyright (c) 2025 Pavlo Moisieienko

package ldi

import (
	"fmt"
	"reflect"
)

type iProvider interface {
	provide(di *Di) (reflect.Value, error)
	getValue() reflect.Value
}

type provideFunction func(iProvider, *Di) (reflect.Value, error)

type providers map[reflect.Type]iProvider

func (ps providers) contains(tp reflect.Type) bool {
	_, ok := ps[tp]
	return ok
}

func (ps providers) getProvider(tp reflect.Type) (iProvider, bool) {
	p, ok := ps[tp]
	return p, ok
}

func (ps providers) addFunction(function any, parameterType reflect.Type, parameterIndex int) error {
	prov := newFunctionProvider(function, parameterIndex)
	return ps.addProvider(parameterType, &prov)
}

func (ps providers) addValue(value reflect.Value) error {
	prov := newValueProvider(value)
	return ps.addProvider(value.Type(), &prov)
}

func (ps providers) addProvider(tp reflect.Type, prov iProvider) error {
	if ps.contains(tp) {
		return fmt.Errorf("provider for type '%s' already exists", tp)
	}
	ps[tp] = prov
	return nil
}

func (ps providers) Len() int {
	return len(ps)
}

func newProviders() providers {
	//revive:disable
	return make(providers, 2)
	//revive:enable
}

type functionProvider struct {
	valueProvider
	function       any
	parameterIndex int
}

func (p *functionProvider) getFunction() any {
	return p.function
}

func (p *functionProvider) getParameterIndex() int {
	return p.parameterIndex
}

func (p *functionProvider) provide(di *Di) (reflect.Value, error) {
	return p.provideFunc(p, di)
}

func newFunctionProvider(function any, parameterIndex int) functionProvider {
	result := functionProvider{
		function:       function,
		parameterIndex: parameterIndex,
	}
	result.provideFunc = func(p iProvider, di *Di) (reflect.Value, error) {
		if p.getValue().IsValid() {
			return p.getValue(), nil
		}
		//revive:disable
		pf := p.(*functionProvider)
		//revive:enable
		values, err := di.innerInvoke(pf.function)
		if err != nil {
			return reflect.Value{}, err
		}
		pf.value = values[pf.parameterIndex]
		return p.getValue(), nil
	}
	return result
}

type valueProvider struct {
	value       reflect.Value
	provideFunc provideFunction
}

func (p *valueProvider) getValue() reflect.Value {
	return p.value
}

func (p *valueProvider) provide(di *Di) (reflect.Value, error) {
	return p.provideFunc(p, di)
}

func newValueProvider(value reflect.Value) valueProvider {
	return valueProvider{
		value: value,
		provideFunc: func(p iProvider, _ *Di) (reflect.Value, error) {
			return p.getValue(), nil
		},
	}
}

func isError(tp reflect.Type) bool {
	return tp.String() == "error"
}

func functionCall(fnc reflect.Value, args []reflect.Value) ([]reflect.Value, error) {
	results := fnc.Call(args)
	for _, r := range results {
		if isError(r.Type()) {
			if err, ok := r.Interface().(error); ok {
				return nil, err
			}
		}
	}
	return results, nil
}
