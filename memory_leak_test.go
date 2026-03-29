// Copyright (c) 2025 Pavlo Moisieienko. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license which can be found in the LICENSE file.

package ldi

import (
	"testing"
)

// TestDi_memory_leak_prevention tests that resolution tracking is properly cleaned up
func TestDi_memory_leak_prevention(t *testing.T) {
	const expectedInt = 4
	const testString = "test"

	di := New()

	// Provide a function that will fail
	err := di.Provide(func() string {
		return testString
	})
	if err != nil {
		t.Fatal(err)
	}

	// Provide another function that depends on the first one
	err = di.Provide(func(s string) int {
		return len(s)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Invoke successfully - should clean up resolution tracking
	err = di.Invoke(func(i int) {
		if i != expectedInt { // len("test") = 4
			t.Errorf("expected %d, got %d", expectedInt, i)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check that resolving map is empty after successful invocation
	di.mu.Lock()
	resolvingCount := len(di.resolving)
	di.mu.Unlock()

	if resolvingCount != 0 {
		t.Errorf("expected resolving map to be empty after successful invocation, but has %d entries", resolvingCount)
	}

	// Test cleanup method explicitly
	di.CleanupResolutionTracking()

	di.mu.Lock()
	resolvingCount = len(di.resolving)
	di.mu.Unlock()

	if resolvingCount != 0 {
		t.Errorf("expected resolving map to be empty after cleanup, but has %d entries", resolvingCount)
	}
}

// TestDi_clear_functionality tests the Clear method
func TestDi_clear_functionality(t *testing.T) {
	const testString = "test string"
	const testInt = 42

	di := New()

	// Add some providers
	err := di.Provide(testString)
	if err != nil {
		t.Fatal(err)
	}
	err = di.Provide(testInt)
	if err != nil {
		t.Fatal(err)
	}

	// Verify providers exist
	err = di.Invoke(func(s string, i int) {
		if s != testString {
			t.Errorf("expected '%s', got '%s'", testString, s)
		}
		if i != testInt {
			t.Errorf("expected %d, got %d", testInt, i)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Clear the container
	di.Clear()

	// Verify that providers are cleared
	err = di.Invoke(func(_ string, _ int) {
		t.Error("should not reach here after clear")
	})
	if err == nil {
		t.Fatal("expected error after clear, but got nil")
	}

	// Check for expected error pattern
	expectedError := "provider for parameter[0] of type 'string' not found"
	if !containsSubstring(err.Error(), expectedError) {
		t.Fatalf("expected error containing '%s', but got: %s", expectedError, err.Error())
	}
}

// Helper function to check if error contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// Simple substring finder
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
