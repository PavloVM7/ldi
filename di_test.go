// Copyright (c) 2025 Pavlo Moisieienko. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license which can be found in the LICENSE file.

package ldi

import (
	"fmt"
	"strings"
	"sync"
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
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected error containing '%s', but got: %s", expectedError, err.Error())
	}
}

// TestDi_concurrent_reads tests multiple goroutines reading from the same DI container
func TestDi_concurrent_reads(t *testing.T) {
	di := New()

	// Provide shared dependencies
	err := di.Provide("shared string")
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(42)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(3.14)
	if err != nil {
		t.Fatal(err)
	}

	// Test concurrent reads
	const numGoroutines = 100
	const numInvocations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numInvocations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numInvocations; j++ {
				err := di.Invoke(func(s string, num int, pi float64) {
					// Verify values are correct
					if s != "shared string" {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 'shared string', got '%s'", id, j, s)
						return
					}
					if num != 42 {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 42, got %d", id, j, num)
						return
					}
					if pi != 3.14 {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 3.14, got %f", id, j, pi)
						return
					}
				})
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, invocation %d: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

// TestDi_concurrent_mixed_operations tests concurrent read and write operations
func TestDi_concurrent_mixed_operations(t *testing.T) {
	di := New()

	// Pre-provide some base dependencies
	err := di.Provide("base string")
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(100)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Half goroutines will provide new values
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine provides a unique type to avoid conflicts
			type uniqueType struct{ id int }
			err := di.Provide(&uniqueType{id: id})
			if err != nil {
				errors <- fmt.Errorf("provide goroutine %d: %v", id, err)
			}
		}(i)
	}

	// Half goroutines will invoke functions
	for i := numGoroutines / 2; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := di.Invoke(func(s string, num int) {
				// Verify base values
				if s != "base string" {
					errors <- fmt.Errorf("invoke goroutine %d: expected 'base string', got '%s'", id, s)
					return
				}
				if num != 100 {
					errors <- fmt.Errorf("invoke goroutine %d: expected 100, got %d", id, num)
					return
				}
			})
			if err != nil {
				errors <- fmt.Errorf("invoke goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

// TestDi_concurrent_function_providers tests concurrent function provider execution
func TestDi_concurrent_function_providers(t *testing.T) {
	di := New()

	// Provide function providers
	err := di.Provide(func() string {
		return "dynamic string"
	})
	if err != nil {
		t.Fatal(err)
	}

	err = di.Provide(func() int {
		return 999
	})
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := di.Invoke(func(s string, num int) {
				// Verify function provider results
				if s != "dynamic string" {
					errors <- fmt.Errorf("goroutine %d: expected 'dynamic string', got '%s'", id, s)
					return
				}
				if num != 999 {
					errors <- fmt.Errorf("goroutine %d: expected 999, got %d", id, num)
					return
				}
			})
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

// TestDi_concurrent_parent_child tests concurrent access to parent-child DI containers
func TestDi_concurrent_parent_child(t *testing.T) {
	parent := New()

	// Provide parent dependencies
	err := parent.Provide("parent string")
	if err != nil {
		t.Fatal(err)
	}
	err = parent.Provide(200)
	if err != nil {
		t.Fatal(err)
	}

	// Create child container
	child := NewWithParent(parent)

	// Provide child-specific dependency
	err = child.Provide("child string")
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Half goroutines use parent container
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := parent.Invoke(func(s string, num int) {
				if s != "parent string" {
					errors <- fmt.Errorf("parent goroutine %d: expected 'parent string', got '%s'", id, s)
					return
				}
				if num != 200 {
					errors <- fmt.Errorf("parent goroutine %d: expected 200, got %d", id, num)
					return
				}
			})
			if err != nil {
				errors <- fmt.Errorf("parent goroutine %d: %v", id, err)
			}
		}(i)
	}

	// Half goroutines use child container
	for i := numGoroutines / 2; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := child.Invoke(func(s string, num int) {
				// Child should resolve to parent's string provider
				if s != "parent string" {
					errors <- fmt.Errorf("child goroutine %d: expected 'parent string', got '%s'", id, s)
					return
				}
				if num != 200 {
					errors <- fmt.Errorf("child goroutine %d: expected 200, got %d", id, num)
					return
				}
			})
			if err != nil {
				errors <- fmt.Errorf("child goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

// TestDi_stress_test_concurrent_access is a stress test for high-concurrency scenarios
func TestDi_stress_test_concurrent_access(t *testing.T) {
	di := New()

	// Provide various types of dependencies
	err := di.Provide("stress test string")
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(12345)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(3.14159)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(true)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 500
	const numInvocations = 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numInvocations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numInvocations; j++ {
				err := di.Invoke(func(s string, num int, pi float64, flag bool) {
					// Verify all values
					if s != "stress test string" {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 'stress test string', got '%s'", id, j, s)
						return
					}
					if num != 12345 {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 12345, got %d", id, j, num)
						return
					}
					if pi != 3.14159 {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected 3.14159, got %f", id, j, pi)
						return
					}
					if !flag {
						errors <- fmt.Errorf("goroutine %d, invocation %d: expected true, got %v", id, j, flag)
						return
					}
				})
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, invocation %d: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
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
