package httpclient

import "net/http"

// Helper function to compare maps
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		// Handle different numeric types that JSON might produce
		if !valuesEqual(valueA, valueB) {
			return false
		}
	}

	return true
}

// Helper function to compare values with type flexibility
func valuesEqual(a, b any) bool {
	if a == b {
		return true
	}

	// Handle numeric conversions
	aFloat, aIsFloat := a.(float64)
	bFloat, bIsFloat := b.(float64)
	if aIsFloat && bIsFloat {
		return aFloat == bFloat
	}

	aInt, aIsInt := a.(int)
	bInt, bIsInt := b.(int)
	if aIsInt && bIsInt {
		return aInt == bInt
	}

	// Cross-type numeric comparison
	if aIsFloat && bIsInt {
		return aFloat == float64(bInt)
	}
	if aIsInt && bIsFloat {
		return float64(aInt) == bFloat
	}

	return false
}

// Helper function to compare any values including slices
func anyEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle slice comparisons
	sliceA, isSliceA := a.([]any)
	sliceB, isSliceB := b.([]any)
	if isSliceA && isSliceB {
		if len(sliceA) != len(sliceB) {
			return false
		}
		for i := range sliceA {
			if !anyEqual(sliceA[i], sliceB[i]) {
				return false
			}
		}
		return true
	}

	// Handle map comparisons
	mapA, isMapA := a.(map[string]any)
	mapB, isMapB := b.(map[string]any)
	if isMapA && isMapB {
		return mapsEqual(mapA, mapB)
	}

	// Handle numeric type conversions
	return valuesEqual(a, b)
}

// Helper types for testing

// failingTransport simulates a transport that always fails
type failingTransport struct{}

func (ft *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, &mockNetError{}
}

// mockNetError simulates a network error
type mockNetError struct{}

func (e *mockNetError) Error() string   { return "mock network error" }
func (e *mockNetError) Timeout() bool   { return false }
func (e *mockNetError) Temporary() bool { return false }
