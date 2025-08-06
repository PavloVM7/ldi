package ldi

import (
	"fmt"
	"reflect"
)

type providers map[reflect.Type]provider

func (ps providers) contains(tp reflect.Type) bool {
	_, ok := ps[tp]
	return ok
}

func (ps providers) getProvider(tp reflect.Type) (provider, bool) {
	p, ok := ps[tp]
	return p, ok
}

func (ps providers) addFunction(function any, parameterType reflect.Type, parameterIndex int) error {
	prov := newFunctionProvider(function, parameterIndex)
	return ps.addProvider(parameterType, prov)
}

func (ps providers) addValue(value reflect.Value) error {
	prov := newValueProvider(value)
	return ps.addProvider(value.Type(), prov)
}

func (ps providers) addProvider(tp reflect.Type, prov provider) error {
	if ps.contains(tp) {
		return newProviderExistsError(tp)
	}
	ps[tp] = prov
	return nil
}

func (ps providers) Len() int {
	return len(ps)
}

func newProviders() providers {
	return make(providers, 2)
}

type provideFunction func(*provider, *Di) (reflect.Value, error)

type provider struct {
	value          reflect.Value
	function       any
	parameterIndex int
	provideFunc    provideFunction
}

func (p *provider) provide(di *Di) (reflect.Value, error) {
	return p.provideFunc(p, di)
}

func newFunctionProvider(function any, parameterIndex int) provider {
	return provider{
		function:       function,
		parameterIndex: parameterIndex,
		provideFunc: func(p *provider, di *Di) (reflect.Value, error) {
			if p.value.IsValid() {
				return p.value, nil
			}
			values, err := di.invoke(p.function)
			if err != nil {
				return reflect.Value{}, err
			}
			p.value = values[p.parameterIndex]
			return p.value, nil
		},
	}
}

func newValueProvider(value reflect.Value) provider {
	return provider{
		value: value,
		provideFunc: func(p *provider, di *Di) (reflect.Value, error) {
			return p.value, nil
		},
		parameterIndex: -1,
	}
}

func newProviderExistsError(tp reflect.Type) error {
	return fmt.Errorf("provider for type '%s' already exists", tp)
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
