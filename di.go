package ldi

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

func New() *Di {
	return NewWithParent(nil)
}
func NewWithParent(parent *Di) *Di {
	return &Di{
		providers: newProviders(),
		parent:    parent,
	}
}

type Di struct {
	providers providers
	parent    *Di
	mu        sync.Mutex
}

func (d *Di) MustProvide(provide any) *Di {
	if err := d.Provide(provide); err != nil {
		log.Fatal(err)
	}
	return d
}

func (d *Di) Provide(anything any) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	val := reflect.ValueOf(anything)
	if val.Kind() == reflect.Func {
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

func (d *Di) invoke(function any) ([]reflect.Value, error) {
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
		paramValue, err := d.provideParameter(d, functionType.In(i), i)
		if err != nil {
			return nil, err
		}
		parameterValues = append(parameterValues, paramValue)
	}
	return functionCall(funcValue, parameterValues)
}

func (d *Di) provideParameter(di *Di, parameterType reflect.Type, parameterIndex int) (reflect.Value, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	prov, ok := d.providers.getProvider(parameterType)
	if !ok {
		if d.parent != nil {
			return d.parent.provideParameter(di, parameterType, parameterIndex)
		}
		return reflect.Value{}, fmt.Errorf("provider for paramter[%d] of type '%s' not found",
			parameterIndex, parameterType)
	}
	return prov.provide(di)
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
