package helpers

import "testing"

func TestConvertMiBToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "zero value",
			input:    0,
			expected: 0,
		},
		{
			name:     "negative value",
			input:    -5,
			expected: -5,
		},
		{
			name:     "1 MiB",
			input:    1,
			expected: 1024 * 1024, // 1,048,576 bytes
		},
		{
			name:     "10 MiB",
			input:    10,
			expected: 10 * 1024 * 1024, // 10,485,760 bytes
		},
		{
			name:     "512 MiB",
			input:    512,
			expected: 512 * 1024 * 1024, // 536,870,912 bytes
		},
		{
			name:     "1024 MiB (1 GiB)",
			input:    1024,
			expected: 1024 * 1024 * 1024, // 1,073,741,824 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMiBToBytes(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertMiBToBytes(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
