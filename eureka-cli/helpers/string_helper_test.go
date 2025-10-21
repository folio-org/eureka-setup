package helpers

import "testing"

func TestTrimModuleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "module name with version suffix",
			input:    "mod-users-123",
			expected: "mod-users",
		},
		{
			name:     "module name with multiple dashes",
			input:    "mod-inventory-storage-25.0.0",
			expected: "mod-inventory-storage",
		},
		{
			name:     "module name without dash",
			input:    "singlename",
			expected: "singlename", // Should handle edge case gracefully
		},
		{
			name:     "module name ending with dash",
			input:    "mod-test-",
			expected: "mod-test",
		},
		{
			name:     "complex module name",
			input:    "mod-data-export-spring-2.0.1",
			expected: "mod-data-export-spring",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimModuleName(tt.input)
			if result != tt.expected {
				t.Errorf("TrimModuleName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
