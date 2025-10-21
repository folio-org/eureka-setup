package helpers

import (
	"bytes"
	"testing"
)

func TestGetVaultRootTokenFromLogs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid vault token log line",
			input:    "Root Token: hvs.abc123def456",
			expected: "hvs.abc123def456",
		},
		{
			name:     "vault token with extra whitespace",
			input:    "Root Token:   hvs.xyz789   ",
			expected: "hvs.xyz789",
		},
		{
			name:     "vault token in longer log line",
			input:    "[INFO] vault: Root Token: hvs.mytoken123 - store this safely",
			expected: "hvs.mytoken123 - store this safely",
		},
		{
			name:     "simple colon pattern",
			input:    "Token: simple_token",
			expected: "simple_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVaultRootTokenFromLogs(tt.input)
			if result != tt.expected {
				t.Errorf("GetVaultRootTokenFromLogs(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPortFromURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:        "valid port URL",
			input:       "http://localhost:8080",
			expected:    8080,
			expectError: false,
		},
		{
			name:        "HTTPS URL with port",
			input:       "https://example.com:9443",
			expected:    9443,
			expectError: false,
		},
		{
			name:        "URL with port only - no path",
			input:       "http://service:3000",
			expected:    3000,
			expectError: false,
		},
		{
			name:        "invalid port - not a number",
			input:       "http://localhost:invalid",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetPortFromURL(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetPortFromURL(%q) expected error, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("GetPortFromURL(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("GetPortFromURL(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetModuleNameFromID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple module ID",
			input:    "mod-users-1.0.0",
			expected: "mod-users",
		},
		{
			name:     "complex module ID",
			input:    "mod-inventory-storage-25.0.1-SNAPSHOT",
			expected: "mod-inventory-storage",
		},
		{
			name:     "module with underscores",
			input:    "mod_data_export_spring-2.0.0",
			expected: "mod_data_export_spring",
		},
		{
			name:     "module with hyphens in name",
			input:    "mod-kb-ebsco-java-3.11.0",
			expected: "mod-kb-ebsco-java",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModuleNameFromID(tt.input)
			if result != tt.expected {
				t.Errorf("GetModuleNameFromID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetModuleVersionFromID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version",
			input:    "mod-users-1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "version with snapshot",
			input:    "mod-inventory-storage-25.0.1-SNAPSHOT",
			expected: "25.0.1-SNAPSHOT",
		},
		{
			name:     "complex version with build metadata",
			input:    "mod-circulation-23.2.0-alpha.1",
			expected: "23.2.0-alpha.1",
		},
		{
			name:     "version with date",
			input:    "mod-kb-ebsco-java-3.11.0-20230915",
			expected: "3.11.0-20230915",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModuleVersionFromID(tt.input)
			if result != tt.expected {
				t.Errorf("GetModuleVersionFromID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetModuleVersionPFromID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version pointer",
			input:    "mod-users-1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "complex version pointer",
			input:    "mod-inventory-storage-25.0.1-SNAPSHOT",
			expected: "25.0.1-SNAPSHOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModuleVersionPFromID(tt.input)
			if result == nil {
				t.Errorf("GetModuleVersionPFromID(%q) returned nil pointer", tt.input)
				return
			}
			if *result != tt.expected {
				t.Errorf("GetModuleVersionPFromID(%q) = %q, want %q", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestGetKafkaConsumerLagFromLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    bytes.Buffer
		expected string
	}{
		{
			name: "log line with newlines",
			input: func() bytes.Buffer {
				var buf bytes.Buffer
				buf.WriteString("consumer-lag: 100\n\r")
				return buf
			}(),
			expected: "consumerlag:100", // Pattern removes [\r\n\s-]+
		},
		{
			name: "log line with multiple newlines and spaces",
			input: func() bytes.Buffer {
				var buf bytes.Buffer
				buf.WriteString("lag: 50\n\r\n  ")
				return buf
			}(),
			expected: "lag:50",
		},
		{
			name: "simple log without newlines",
			input: func() bytes.Buffer {
				var buf bytes.Buffer
				buf.WriteString("no-lag")
				return buf
			}(),
			expected: "nolag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetKafkaConsumerLagFromLogLine(tt.input)
			if result != tt.expected {
				t.Errorf("GetKafkaConsumerLagFromLogLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMatchesModuleName(t *testing.T) {
	tests := []struct {
		name       string
		moduleID   string
		moduleName string
		expected   bool
	}{
		{
			name:       "exact match with version",
			moduleID:   "mod-users-1.0.0",
			moduleName: "mod-users",
			expected:   true,
		},
		{
			name:       "match with complex version",
			moduleID:   "mod-inventory-storage-25.0.1-SNAPSHOT",
			moduleName: "mod-inventory-storage",
			expected:   true,
		},
		{
			name:       "no match - different module",
			moduleID:   "mod-circulation-1.0.0",
			moduleName: "mod-users",
			expected:   false,
		},
		{
			name:       "no match - partial module name",
			moduleID:   "mod-inventory-storage-1.0.0",
			moduleName: "mod-inventory",
			expected:   false,
		},
		{
			name:       "no match - module name too long",
			moduleID:   "mod-users-1.0.0",
			moduleName: "mod-users-extra",
			expected:   false,
		},
		{
			name:       "match with single digit version",
			moduleID:   "mod-test-1",
			moduleName: "mod-test",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesModuleName(tt.moduleID, tt.moduleName)
			if result != tt.expected {
				t.Errorf("MatchesModuleName(%q, %q) = %t, want %t", tt.moduleID, tt.moduleName, result, tt.expected)
			}
		})
	}
}
