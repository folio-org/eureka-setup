package helpers

import "testing"

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		key      string
		expected bool
	}{
		{
			name: "key exists with true value",
			input: map[string]any{
				"enabled": true,
			},
			key:      "enabled",
			expected: true,
		},
		{
			name: "key exists with false value",
			input: map[string]any{
				"enabled": false,
			},
			key:      "enabled",
			expected: false,
		},
		{
			name: "key does not exist",
			input: map[string]any{
				"other": true,
			},
			key:      "enabled",
			expected: false,
		},
		{
			name: "key exists with nil value",
			input: map[string]any{
				"enabled": nil,
			},
			key:      "enabled",
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			key:      "enabled",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBool(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("GetBool(%v, %q) = %t, want %t", tt.input, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetAnyOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]any
		key          string
		defaultValue any
		expected     any
	}{
		{
			name: "key exists",
			input: map[string]any{
				"value": "test",
			},
			key:          "value",
			defaultValue: "default",
			expected:     "test",
		},
		{
			name: "key does not exist",
			input: map[string]any{
				"other": "test",
			},
			key:          "value",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name: "key exists with nil value",
			input: map[string]any{
				"value": nil,
			},
			key:          "value",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty map",
			input:        map[string]any{},
			key:          "value",
			defaultValue: 42,
			expected:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAnyOrDefault(tt.input, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetAnyOrDefault(%v, %q, %v) = %v, want %v", tt.input, tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]any
		key          string
		defaultValue int64
		expected     int64
	}{
		{
			name: "key exists with int value",
			input: map[string]any{
				"count": 42,
			},
			key:          "count",
			defaultValue: 10,
			expected:     42,
		},
		{
			name: "key does not exist",
			input: map[string]any{
				"other": 42,
			},
			key:          "count",
			defaultValue: 10,
			expected:     10,
		},
		{
			name: "key exists with nil value",
			input: map[string]any{
				"count": nil,
			},
			key:          "count",
			defaultValue: 10,
			expected:     10,
		},
		{
			name: "key exists with wrong type",
			input: map[string]any{
				"count": "not-a-number",
			},
			key:          "count",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "empty map",
			input:        map[string]any{},
			key:          "count",
			defaultValue: 100,
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIntOrDefault(tt.input, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetIntOrDefault(%v, %q, %d) = %d, want %d", tt.input, tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]any
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name: "key exists with true value",
			input: map[string]any{
				"enabled": true,
			},
			key:          "enabled",
			defaultValue: false,
			expected:     true,
		},
		{
			name: "key exists with false value",
			input: map[string]any{
				"enabled": false,
			},
			key:          "enabled",
			defaultValue: true,
			expected:     false,
		},
		{
			name: "key does not exist",
			input: map[string]any{
				"other": true,
			},
			key:          "enabled",
			defaultValue: true,
			expected:     true,
		},
		{
			name: "key exists with nil value",
			input: map[string]any{
				"enabled": nil,
			},
			key:          "enabled",
			defaultValue: true,
			expected:     true,
		},
		{
			name: "key exists with wrong type",
			input: map[string]any{
				"enabled": "not-a-bool",
			},
			key:          "enabled",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "empty map",
			input:        map[string]any{},
			key:          "enabled",
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolOrDefault(tt.input, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetBoolOrDefault(%v, %q, %t) = %t, want %t", tt.input, tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
