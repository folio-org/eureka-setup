package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

func TestReadJsonFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid JSON file",
			fileContent: `{"name": "test", "value": 123}`,
			filePath:    "test_valid.json",
			expectError: false,
		},
		{
			name:        "invalid JSON file",
			fileContent: `{"name": "test", "value":}`,
			filePath:    "test_invalid.json",
			expectError: true,
		},
		{
			name:        "empty JSON file",
			fileContent: `{}`,
			filePath:    "test_empty.json",
			expectError: false,
		},
		{
			name:        "non-existent file",
			fileContent: "",
			filePath:    "non_existent.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.name != "non-existent file" {
				// Create temporary file
				tmpDir := t.TempDir()
				filePath = filepath.Join(tmpDir, tt.filePath)
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			} else {
				filePath = tt.filePath
			}

			action := &action.Action{Name: "test"}
			var result map[string]interface{}
			err := ReadJsonFromFile(action, filePath, &result)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify result is parsed correctly for valid JSON
				if tt.name == "valid JSON file" {
					if result["name"] != "test" {
						t.Errorf("Expected name=test, got %v", result["name"])
					}
					if result["value"] != float64(123) { // JSON numbers are float64
						t.Errorf("Expected value=123, got %v", result["value"])
					}
				}
			}
		})
	}
}

func TestWriteJsonToFile(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		expectError bool
		setupFile   bool
	}{
		{
			name: "valid struct",
			data: map[string]interface{}{
				"name":  "test",
				"value": 123,
			},
			expectError: false,
			setupFile:   true,
		},
		{
			name:        "simple string",
			data:        "simple string",
			expectError: false,
			setupFile:   true,
		},
		{
			name: "complex nested structure",
			data: map[string]interface{}{
				"nested": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": "value",
					},
				},
				"array": []string{"item1", "item2"},
			},
			expectError: false,
			setupFile:   true,
		},
		{
			name:        "channel type (unmarshalable)",
			data:        make(chan int),
			expectError: true,
			setupFile:   true,
		},
		{
			name:        "write to non-existent file",
			data:        map[string]interface{}{"test": "data"},
			expectError: true,
			setupFile:   false, // Don't create the initial file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.json")
			action := &action.Action{Name: "test"}

			if tt.setupFile {
				// Create the file first as WriteJsonToFile expects it to exist
				err := os.WriteFile(filePath, []byte("{}"), 0644)
				if err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
			}

			err := WriteJsonToFile(action, filePath, tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file exists and contains valid JSON
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Error("Expected file to exist")
				}

				// Read back and verify JSON is valid
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read file content: %v", err)
				}

				var parsed interface{}
				err = json.Unmarshal(content, &parsed)
				if err != nil {
					t.Errorf("Written file contains invalid JSON: %v", err)
				}

				// Verify the JSON is properly formatted (indented)
				if len(content) > 10 { // Only check for non-trivial JSON
					contentStr := string(content)
					if !strings.Contains(contentStr, "\n") {
						t.Error("JSON should be formatted with indentation")
					}
				}
			}
		})
	}
}

func TestCheckIsRegularFile(t *testing.T) {
	tests := []struct {
		name        string
		fileType    string // "regular", "directory", "nonexistent"
		expectError bool
	}{
		{
			name:        "regular file",
			fileType:    "regular",
			expectError: false,
		},
		{
			name:        "directory",
			fileType:    "directory",
			expectError: true,
		},
		{
			name:        "non-existent file",
			fileType:    "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var filePath string
			action := &action.Action{Name: "test"}

			switch tt.fileType {
			case "regular":
				filePath = filepath.Join(tmpDir, "regular.txt")
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			case "directory":
				filePath = filepath.Join(tmpDir, "testdir")
				err := os.Mkdir(filePath, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			case "nonexistent":
				filePath = filepath.Join(tmpDir, "nonexistent.txt")
			}

			err := CheckIsRegularFile(action, filePath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCopySingleFile(t *testing.T) {
	tests := []struct {
		name         string
		srcContent   string
		createSrc    bool
		createDstDir bool
		expectError  bool
	}{
		{
			name:         "copy existing file",
			srcContent:   "test file content",
			createSrc:    true,
			createDstDir: true,
			expectError:  false,
		},
		{
			name:         "copy non-existent file",
			srcContent:   "",
			createSrc:    false,
			createDstDir: true,
			expectError:  true,
		},
		{
			name:         "copy large file",
			srcContent:   string(make([]byte, 10000)),
			createSrc:    true,
			createDstDir: true,
			expectError:  false,
		},
		{
			name:         "copy to non-existent directory",
			srcContent:   "test content",
			createSrc:    true,
			createDstDir: false,
			expectError:  true,
		},
		{
			name:         "copy empty file",
			srcContent:   "",
			createSrc:    true,
			createDstDir: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcPath := filepath.Join(tmpDir, "source.txt")

			var dstPath string
			if tt.createDstDir {
				dstPath = filepath.Join(tmpDir, "destination.txt")
			} else {
				// Use a non-existent directory
				dstPath = filepath.Join(tmpDir, "nonexistent", "destination.txt")
			}

			action := &action.Action{Name: "test"}

			if tt.createSrc {
				err := os.WriteFile(srcPath, []byte(tt.srcContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}
			}

			err := CopySingleFile(action, srcPath, dstPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify destination file exists and has correct content
				dstContent, err := os.ReadFile(dstPath)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
				}

				if string(dstContent) != tt.srcContent {
					t.Errorf("Destination file content mismatch: got %q, want %q", string(dstContent), tt.srcContent)
				}

				// Verify file permissions
				srcInfo, err := os.Stat(srcPath)
				if err != nil {
					t.Errorf("Failed to stat source file: %v", err)
				}

				dstInfo, err := os.Stat(dstPath)
				if err != nil {
					t.Errorf("Failed to stat destination file: %v", err)
				}

				if srcInfo.Mode() != dstInfo.Mode() {
					t.Errorf("File permissions not preserved: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
				}
			}
		})
	}
}

func TestGetCurrentWorkDirPath(t *testing.T) {
	action := &action.Action{Name: "test"}

	workDir, err := GetCurrentWorkDirPath(action)
	if err != nil {
		t.Errorf("GetCurrentWorkDirPath failed: %v", err)
	}

	if workDir == "" {
		t.Error("GetCurrentWorkDirPath returned empty string")
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(workDir) {
		t.Errorf("GetCurrentWorkDirPath should return absolute path, got: %s", workDir)
	}

	// Verify the directory actually exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		t.Errorf("Current working directory does not exist: %s", workDir)
	}
}

func TestGetHomeMiscDir(t *testing.T) {
	action := &action.Action{Name: "test"}

	miscDir, err := GetHomeMiscDir(action)
	if err != nil {
		t.Errorf("GetHomeMiscDir failed: %v", err)
	}

	if miscDir == "" {
		t.Error("GetHomeMiscDir returned empty string")
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(miscDir) {
		t.Errorf("GetHomeMiscDir should return absolute path, got: %s", miscDir)
	}

	// Verify it contains the expected subdirectory structure
	expectedSuffix := filepath.Join(".eureka", "misc")
	if !strings.HasSuffix(miscDir, expectedSuffix) {
		t.Errorf("GetHomeMiscDir should contain %s, got: %s", expectedSuffix, miscDir)
	}
}

func TestGetHomeDirPath(t *testing.T) {
	action := &action.Action{Name: "test"}

	homeDir, err := GetHomeDirPath(action)
	if err != nil {
		t.Errorf("GetHomeDirPath failed: %v", err)
	}

	if homeDir == "" {
		t.Error("GetHomeDirPath returned empty string")
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(homeDir) {
		t.Errorf("GetHomeDirPath should return absolute path, got: %s", homeDir)
	}

	// Verify the directory was created and exists
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		t.Errorf("Home directory should have been created: %s", homeDir)
	}

	// Verify it contains the expected subdirectory structure
	if !strings.HasSuffix(homeDir, ".eureka") {
		t.Errorf("GetHomeDirPath should end with '.eureka', got: %s", homeDir)
	}

	// Test that calling it again returns the same directory
	homeDir2, err2 := GetHomeDirPath(action)
	if err2 != nil {
		t.Errorf("Second call to GetHomeDirPath failed: %v", err2)
	}

	if homeDir != homeDir2 {
		t.Errorf("GetHomeDirPath should return consistent results: first=%s, second=%s", homeDir, homeDir2)
	}
}

func TestCloseFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CloseFile panicked: %v", r)
		}
	}()

	CloseFile(file)

	// Try to write to closed file - should fail
	_, err = file.WriteString("test")
	if err == nil {
		t.Error("Expected error when writing to closed file")
	}
}

func TestCloseReader(t *testing.T) {
	// Create a test file and open it as reader
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	err := os.WriteFile(filePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CloseReader panicked: %v", r)
		}
	}()

	CloseReader(file)

	// Try to read from closed file - should fail
	_, err = file.Read(make([]byte, 10))
	if err == nil {
		t.Error("Expected error when reading from closed file")
	}
}
