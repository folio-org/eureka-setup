package helpers

import "testing"

func TestConvertMapKeysToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: []string{},
		},
		{
			name: "single key map",
			input: map[string]any{
				"key1": "value1",
			},
			expected: []string{"key1"},
		},
		{
			name: "multiple keys map",
			input: map[string]any{
				"mod-users":     "1.0.0",
				"mod-inventory": "2.0.0",
				"mod-circulation": "3.0.0",
			},
			expected: []string{"mod-users", "mod-inventory", "mod-circulation"},
		},
		{
			name: "map with mixed value types",
			input: map[string]any{
				"string": "value",
				"number": 42,
				"bool":   true,
				"nil":    nil,
			},
			expected: []string{"string", "number", "bool", "nil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMapKeysToSlice(tt.input)
			
			// Since map iteration order is not guaranteed, we need to compare as sets
			if len(result) != len(tt.expected) {
				t.Errorf("ConvertMapKeysToSlice() returned %d keys, want %d", len(result), len(tt.expected))
				return
			}
			
			// Convert to map for easier comparison
			resultMap := make(map[string]bool)
			for _, key := range result {
				resultMap[key] = true
			}
			
			for _, expectedKey := range tt.expected {
				if !resultMap[expectedKey] {
					t.Errorf("ConvertMapKeysToSlice() missing expected key %q", expectedKey)
				}
			}
		})
	}
}