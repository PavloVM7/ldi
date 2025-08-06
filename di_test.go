package ldi

import (
	"testing"
)

func TestDi_Provide(t *testing.T) {
	di := New()
	err := di.Provide(GetITest)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide("test string")
	if err != nil {
		t.Fatal(err)
	}
	if di.providers.Len() != 2 {
		t.Fatalf("expected 1 provider, but got %d", di.providers.Len())
	}
}
func TestDi_Provide_value_int_duplicate(t *testing.T) {
	di := New()
	err := di.Provide(1)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(2)
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	if di.providers.Len() != 1 {
		t.Fatalf("expected 1 provider, but got %d", di.providers.Len())
	}
}
func TestDi_Provide_value(t *testing.T) {
	di := New()
	err := di.Provide("string")
	if err != nil {
		t.Fatal(err)
	}
	if di.providers.Len() != 1 {
		t.Fatalf("expected 1 provider, but got %d", di.providers.Len())
	}
}

func TestDi_Provide_function_error_not_provided(t *testing.T) {
	di := New()
	err := di.Provide(func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if di.providers.Len() != 0 {
		t.Fatalf("expected 0 provider, but got %d", di.providers.Len())
	}
}
func TestDi_Provide_function_no_return_value(t *testing.T) {
	di := New()
	err := di.Provide(func1)
	if err == nil {
		t.Fatal(err)
	}
	err = di.Provide(func() {})
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	if di.providers.Len() != 0 {
		t.Fatalf("expected 0 provider, but got %d", di.providers.Len())
	}
}
func TestDi_Provide_function_duplicate(t *testing.T) {
	di := New()
	err := di.Provide(func() ITest {
		return &testStruct{strValue: "inner function"}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(func() ITest {
		return &testStruct{strValue: "inner function 2"}
	})
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	err = di.Provide(GetITest)
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	if di.providers.Len() != 1 {
		t.Fatalf("expected 1 provider, but got %d", di.providers.Len())
	}
}
func TestDi_Provide_function(t *testing.T) {
	di := New()
	err := di.Provide(GetITest)
	if err != nil {
		t.Fatal(err)
	}
	if di.providers.Len() != 1 {
		t.Fatalf("expected 1 provider, but got %d", di.providers.Len())
	}
}

type ITest interface {
	GetString() string
}

func GetITest(s string) ITest {
	return &testStruct{strValue: s}
}
func func1() {
	// do nothing
}

type testStruct struct {
	strValue string
}

func (t *testStruct) GetString() string {
	return t.strValue
}
