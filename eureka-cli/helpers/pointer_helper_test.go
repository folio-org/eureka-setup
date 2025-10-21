package helpers

import "testing"

func TestStringP(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "normal string",
			input: "test-value",
		},
		{
			name:  "module name",
			input: "mod-users-1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringP(tt.input)
			if result == nil {
				t.Error("StringP() returned nil pointer")
				return
			}
			if *result != tt.input {
				t.Errorf("StringP(%q) = %q, want %q", tt.input, *result, tt.input)
			}
		})
	}
}

func TestBoolP(t *testing.T) {
	tests := []struct {
		name  string
		input bool
	}{
		{
			name:  "true value",
			input: true,
		},
		{
			name:  "false value",
			input: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolP(tt.input)
			if result == nil {
				t.Error("BoolP() returned nil pointer")
				return
			}
			if *result != tt.input {
				t.Errorf("BoolP(%t) = %t, want %t", tt.input, *result, tt.input)
			}
		})
	}
}

func TestIntP(t *testing.T) {
	tests := []struct {
		name  string
		input int
	}{
		{
			name:  "zero value",
			input: 0,
		},
		{
			name:  "positive value",
			input: 42,
		},
		{
			name:  "negative value",
			input: -10,
		},
		{
			name:  "large value",
			input: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntP(tt.input)
			if result == nil {
				t.Error("IntP() returned nil pointer")
				return
			}
			if *result != tt.input {
				t.Errorf("IntP(%d) = %d, want %d", tt.input, *result, tt.input)
			}
		})
	}
}
