package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// CreateTempJSONFile creates a temporary JSON file with the given data
func CreateTempJSONFile(t *testing.T, data any) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-descriptor.json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(tmpFile, jsonData, 0600)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return tmpFile
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-file.txt")

	err := os.WriteFile(tmpFile, []byte(content), 0600)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return tmpFile
}

// CreateJSONFileInDir creates a JSON file with the given name in the specified directory
func CreateJSONFileInDir(t *testing.T, dir string, filename string, data any) string {
	t.Helper()

	filePath := filepath.Join(dir, filename)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(filePath, jsonData, 0600)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	return filePath
}
