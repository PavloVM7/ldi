// Copyright (c) 2025 Pavlo Moisieienko. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license which can be found in the LICENSE file.

package ldi

import (
	"fmt"
	"testing"
)

func TestDi_Invoke_private(t *testing.T) {
	parent1 := New().MustProvide("private test string")
	di := NewWithParent(parent1).MustProvide(getIPrivate)
	err := di.Invoke(func(ts iPrivate) {
		if ts.GetStringValue() != "private test string" {
			t.Fatalf("expected 'private test string', but got '%s'", ts.GetStringValue())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDi_Invoke_two_parents(t *testing.T) {
	parent1 := New().MustProvide("test string")
	parent2 := NewWithParent(parent1).MustProvide(123)
	di := NewWithParent(parent2).MustProvide(GetITest)
	err := di.Invoke(func(ts ITest, vInt int) {
		if vInt != 123 {
			t.Fatalf("expected 123, but got %d", vInt)
		}
		if ts.GetString() != "test string" {
			t.Fatalf("expected 'test string', but got '%s'", ts.GetString())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDi_Invoke_parent(t *testing.T) {
	parent1 := New().MustProvide("test string").MustProvide(123)
	di := NewWithParent(parent1).MustProvide(GetITest)
	err := di.Invoke(func(ts ITest, vInt int) {
		if vInt != 123 {
			t.Fatalf("expected 123, but got %d", vInt)
		}
		if ts.GetString() != "test string" {
			t.Fatalf("expected 'test string', but got '%s'", ts.GetString())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDi_Invoke_function(t *testing.T) {
	di := New().MustProvide(GetITest).MustProvide("test string")
	err := di.Invoke(func(ts ITest) {
		if ts.GetString() != "test string" {
			t.Fatalf("expected 'test string', but got '%s'", ts.GetString())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDi_Invoke(t *testing.T) {
	di := New().MustProvide("test string")
	err := di.Invoke(GetITest)
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestDi_MustProvide_nil_nil(t *testing.T) {
	f := func() ([]int, []string) {
		return nil, nil
	}

	di := New()
	err := di.MustProvide(f).Invoke(func(ints []int, strs []string) {
		if ints != nil {
			t.Fatalf("expected nil, but got %v", ints)
		}
		if strs != nil {
			t.Fatalf("expected nil, but got %v", strs)
		}
	})
	if err != nil {
		t.Fatalf("expected nil, but got %v", err)
	}
}

func TestDi_MustInvoke_nil_slice(t *testing.T) {
	f := func() ([]int, []string) {
		return nil, nil
	}
	di := New().MustProvide(f)
	err := di.Invoke(func(ints []int) {
		if ints != nil {
			t.Fatalf("expected nil, but got %v", ints)
		}
	})
	if err != nil {
		t.Fatalf("expected nil, but got %v", err)
	}
}

func TestDi_function_many_results_parent(t *testing.T) {
	expectedInt := []int{1, 2, 3}
	expectedStr := []string{"a", "b", "c"}
	count := 0
	f := func() ([]int, []string) {
		count++
		return expectedInt, expectedStr
	}
	var (
		actualInt1 []int
		actualStr1 []string
		actualInt2 []int
		actualStr2 []string
	)

	parent := New().MustProvide(f)
	di := NewWithParent(parent).
		MustInvoke(func(ints []int, strings []string) {
			actualInt1 = ints
			actualStr1 = strings
		}).
		MustInvoke(func(ints []int) {
			actualInt2 = ints
		}).MustInvoke(func(strings []string) {
		actualStr2 = strings
	})
	if di == nil {
		t.Fatal("di expected non nil, but got nil")
	}
	if di.parent == nil {
		t.Fatal("parent expected non nil, but got nil")
	}
	arrayEquals(t, actualInt1, expectedInt)
	arrayEquals(t, actualStr1, expectedStr)
	arrayEquals(t, actualInt2, expectedInt)
	arrayEquals(t, actualStr2, expectedStr)
	if count != 1 {
		t.Fatal("expected count to be 1, but got:", count)
	}
}
func TestDi_function_many_results(t *testing.T) {
	expectedInt := []int{1, 2, 3}
	expectedStr := []string{"a", "b", "c"}
	count := 0
	f := func() ([]int, []string) {
		count++
		return expectedInt, expectedStr
	}
	var (
		actualInt1 []int
		actualStr1 []string
		actualInt2 []int
		actualStr2 []string
	)
	New().MustProvide(f).
		MustInvoke(func(ints []int, strings []string) {
			actualInt1 = ints
			actualStr1 = strings
		}).
		MustInvoke(func(ints []int) {
			actualInt2 = ints
		}).MustInvoke(func(strings []string) {
		actualStr2 = strings
	})

	arrayEquals(t, actualInt1, expectedInt)
	arrayEquals(t, actualStr1, expectedStr)
	arrayEquals(t, actualInt2, expectedInt)
	arrayEquals(t, actualStr2, expectedStr)
	if count != 1 {
		t.Fatal("expected count to be 1, but got:", count)
	}
}

func TestDi_no_providers(t *testing.T) {
	parent := New()
	di := NewWithParent(parent).MustProvide(GetITest).MustProvide("some string")
	err := di.Invoke(func(iTest ITest) error {
		if iTest == nil {
			return fmt.Errorf("expected non nil, but got nil")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	pln := di.parent.providers.Len()
	if pln != 0 {
		t.Fatalf("expected 0 parent providers, but got %d", pln)
	}
}

func TestDi_provide_function_nil(t *testing.T) {
	di := New()
	t.Run("provide nil", func(t *testing.T) {
		err := di.Provide(nil)
		t.Log("err:", err)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
	})
	t.Run("provide function nil", func(t *testing.T) {
		err := di.Provide((func() []int)(nil))
		t.Log("err:", err)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
	})
}

type InterfaceA interface{ DoA() }
type InterfaceB interface{ DoB() }

type testA struct {
	b InterfaceB
}

func (t *testA) DoA() {}

type testB struct {
	a InterfaceA
}

func (t *testB) DoB() {}

func TestDi_circular_dependency_detection(t *testing.T) {
	di := New()

	// Provide function that creates A requiring B
	err := di.Provide(func(b InterfaceB) InterfaceA {
		return &testA{b: b}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Provide function that creates B requiring A (circular)
	err = di.Provide(func(a InterfaceA) InterfaceB {
		return &testB{a: a}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Try to resolve A - should detect circular dependency
	err = di.Invoke(func(a InterfaceA) {
		t.Fatal("should not reach here due to circular dependency")
	})

	if err == nil {
		t.Fatal("expected circular dependency error, but got nil")
	}

	expectedError := "circular dependency detected for type"
	if !contains(err.Error(), expectedError) {
		t.Fatalf("expected error containing '%s', but got: %s", expectedError, err.Error())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasPrefix(s, substr) || hasSuffix(s, substr) || containsMiddle(s, substr)))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func arrayEquals[T comparable](t *testing.T, actual []T, expected []T) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatal("expected array length to be equal", len(expected), "!=", len(actual))
	}
	for i := 0; i < len(actual); i++ {
		if actual[i] != expected[i] {
			t.Fatal("expected array value to be equal", expected[i], "!=", actual[i])
		}
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

func getIPrivate(s string) iPrivate {
	return &privateStruct{strValue: s}
}

type iPrivate interface {
	GetStringValue() string
}

type privateStruct struct {
	strValue string
}

func (t *privateStruct) GetStringValue() string {
	return t.strValue
}
