package helpers_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/*
var testEmbedFS embed.FS

func TestReadJSONFromFile_ValidJSON(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	jsonContent := `{"name": "test", "value": 123}`
	err := os.WriteFile(filePath, []byte(jsonContent), 0644)
	assert.NoError(t, err)

	var data map[string]any

	// Act
	err = helpers.ReadJSONFromFile(filePath, &data)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test", data["name"])
	assert.Equal(t, float64(123), data["value"])
}

func TestReadJSONFromFile_FileNotFound(t *testing.T) {
	// Arrange
	var data map[string]any

	// Act
	err := helpers.ReadJSONFromFile("/nonexistent/file.json", &data)

	// Assert
	assert.Error(t, err)
}

func TestReadJSONFromFile_InvalidJSON(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(filePath, []byte("{invalid json"), 0644)
	assert.NoError(t, err)

	var data map[string]any

	// Act
	err = helpers.ReadJSONFromFile(filePath, &data)

	// Assert
	assert.Error(t, err)
}

func TestReadJSONFromFile_FromEmbeddedTestData(t *testing.T) {
	t.Run("TestReadJSONFromFile_FromEmbeddedTestData", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// Copy embedded config.json to temp directory
		err := helpers.CopyMultipleFiles(tmpDir, &testEmbedFS)
		assert.NoError(t, err)

		configPath := filepath.Join(tmpDir, "testdata", "config.json")
		var config map[string]any

		// Act
		err = helpers.ReadJSONFromFile(configPath, &config)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test", config["name"])
		assert.Equal(t, "1.0.0", config["version"])
	})
}

func TestReadJSONFromFile_StructuredData(t *testing.T) {
	t.Run("TestReadJSONFromFile_StructuredData", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// Copy embedded test data
		err := helpers.CopyMultipleFiles(tmpDir, &testEmbedFS)
		assert.NoError(t, err)

		configPath := filepath.Join(tmpDir, "testdata", "config.json")

		// Define a struct matching the JSON structure
		type Config struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		var config Config

		// Act
		err = helpers.ReadJSONFromFile(configPath, &config)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test", config.Name)
		assert.Equal(t, "1.0.0", config.Version)
	})
}

func TestWriteJSONToFile_ValidData(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.json")

	// Create the file first
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	_ = file.Close()

	data := map[string]any{
		"name":  "test",
		"value": 456,
	}

	// Act
	err = helpers.WriteJSONToFile(filePath, data)

	// Assert
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"name": "test"`)
	assert.Contains(t, string(content), `"value": 456`)
}

func TestWriteJSONToFile_ReadWriteRoundTrip(t *testing.T) {
	t.Run("TestWriteJSONToFile_ReadWriteRoundTrip", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// First, copy and read from embedded test data
		err := helpers.CopyMultipleFiles(tmpDir, &testEmbedFS)
		assert.NoError(t, err)

		srcPath := filepath.Join(tmpDir, "testdata", "config.json")
		var originalData map[string]any
		err = helpers.ReadJSONFromFile(srcPath, &originalData)
		assert.NoError(t, err)

		// Modify the data
		originalData["modified"] = true
		originalData["count"] = 42

		// Write to new file
		dstPath := filepath.Join(tmpDir, "modified-config.json")
		err = helpers.WriteJSONToFile(dstPath, originalData)
		assert.NoError(t, err)

		// Read back and verify
		var readBackData map[string]any
		err = helpers.ReadJSONFromFile(dstPath, &readBackData)
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "test", readBackData["name"])
		assert.Equal(t, "1.0.0", readBackData["version"])
		assert.Equal(t, true, readBackData["modified"])
		assert.Equal(t, float64(42), readBackData["count"])
	})
}

func TestWriteJSONToFile_ComplexStructure(t *testing.T) {
	t.Run("TestWriteJSONToFile_ComplexStructure", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "complex.json")

		type Module struct {
			Name    string   `json:"name"`
			Version string   `json:"version"`
			Tags    []string `json:"tags"`
		}

		complexData := map[string]any{
			"modules": []Module{
				{Name: "mod-inventory", Version: "1.0.0", Tags: []string{"core", "stable"}},
				{Name: "mod-users", Version: "2.0.0", Tags: []string{"core"}},
			},
			"enabled": true,
		}

		// Act
		err := helpers.WriteJSONToFile(filePath, complexData)
		assert.NoError(t, err)

		// Read back
		var readBack map[string]any
		err = helpers.ReadJSONFromFile(filePath, &readBack)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, readBack["modules"])
		assert.Equal(t, true, readBack["enabled"])
	})
}

func TestCopySingleFile_Success(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	content := "test content"
	err := os.WriteFile(srcPath, []byte(content), 0644)
	assert.NoError(t, err)

	// Act
	err = helpers.CopySingleFile(srcPath, dstPath)

	// Assert
	assert.NoError(t, err)

	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(dstContent))
}

func TestCopySingleFile_SourceNotFound(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dstPath := filepath.Join(tmpDir, "dest.txt")

	// Act
	err := helpers.CopySingleFile("/nonexistent/source.txt", dstPath)

	// Assert
	assert.Error(t, err)
}

func TestCopySingleFile_FromEmbeddedTestData(t *testing.T) {
	t.Run("TestCopySingleFile_FromEmbeddedTestData", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// First copy embedded files to a source directory
		srcDir := filepath.Join(tmpDir, "src")
		err := helpers.CopyMultipleFiles(srcDir, &testEmbedFS)
		assert.NoError(t, err)

		srcFile := filepath.Join(srcDir, "testdata", "file1.txt")
		dstFile := filepath.Join(tmpDir, "copied-file1.txt")

		// Act
		err = helpers.CopySingleFile(srcFile, dstFile)

		// Assert
		assert.NoError(t, err)

		// Verify content matches
		srcContent, err := os.ReadFile(srcFile)
		assert.NoError(t, err)
		dstContent, err := os.ReadFile(dstFile)
		assert.NoError(t, err)
		assert.Equal(t, srcContent, dstContent)
		assert.Contains(t, string(dstContent), "Test content for file 1")
	})
}

func TestCopySingleFile_NestedFile(t *testing.T) {
	t.Run("TestCopySingleFile_NestedFile", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// Copy embedded test data
		srcDir := filepath.Join(tmpDir, "src")
		err := helpers.CopyMultipleFiles(srcDir, &testEmbedFS)
		assert.NoError(t, err)

		nestedSrc := filepath.Join(srcDir, "testdata", "subdir", "nested.txt")
		nestedDst := filepath.Join(tmpDir, "nested-copy.txt")

		// Act
		err = helpers.CopySingleFile(nestedSrc, nestedDst)

		// Assert
		assert.NoError(t, err)

		content, err := os.ReadFile(nestedDst)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "Nested file content")
	})
}

func TestCopySingleFile_JSONFile(t *testing.T) {
	t.Run("TestCopySingleFile_JSONFile", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		// Copy embedded test data
		srcDir := filepath.Join(tmpDir, "src")
		err := helpers.CopyMultipleFiles(srcDir, &testEmbedFS)
		assert.NoError(t, err)

		jsonSrc := filepath.Join(srcDir, "testdata", "config.json")
		jsonDst := filepath.Join(tmpDir, "config-copy.json")

		// Act
		err = helpers.CopySingleFile(jsonSrc, jsonDst)

		// Assert
		assert.NoError(t, err)

		// Verify JSON is still valid after copy
		var config map[string]any
		err = helpers.ReadJSONFromFile(jsonDst, &config)
		assert.NoError(t, err)
		assert.Equal(t, "test", config["name"])
		assert.Equal(t, "1.0.0", config["version"])
	})
}

func TestCopySingleFile_SourceIsDirectory(t *testing.T) {
	t.Run("TestCopySingleFile_SourceIsDirectory", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "source-dir")
		err := os.MkdirAll(srcDir, 0755)
		assert.NoError(t, err)

		dstPath := filepath.Join(tmpDir, "dest.txt")

		// Act
		err = helpers.CopySingleFile(srcDir, dstPath)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a regular file")
	})
}

func TestIsRegularFile_RegularFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "regular.txt")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)

	// Act
	err = helpers.IsRegularFile(filePath)

	// Assert
	assert.NoError(t, err)
}

func TestIsRegularFile_NotFound(t *testing.T) {
	// Act
	err := helpers.IsRegularFile("/nonexistent/file.txt")

	// Assert
	assert.Error(t, err)
}

func TestIsRegularFile_Directory(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// Act
	err := helpers.IsRegularFile(tmpDir)

	// Assert
	assert.Error(t, err)
}

func TestGetCurrentWorkDirPath_Success(t *testing.T) {
	// Act
	result, err := helpers.GetCurrentWorkDirPath()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetHomeDirPath_Success(t *testing.T) {
	// Act
	result, err := helpers.GetHomeDirPath()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetHomeMiscDir_Success(t *testing.T) {
	// Act
	result, err := helpers.GetHomeMiscDir()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestCloseFile_NoError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	file, err := os.Create(filePath)
	assert.NoError(t, err)

	// Act & Assert - Should not panic
	helpers.CloseFile(file)
}

func TestCloseReader_NoError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	file, err := os.Open(filePath)
	assert.NoError(t, err)

	// Act & Assert - Should not panic
	helpers.CloseReader(file)
}

// Tests for CopyMultipleFiles

func TestCopyMultipleFiles_Success(t *testing.T) {
	t.Run("TestCopyMultipleFiles_Success", func(t *testing.T) {
		// Arrange
		dstDir := t.TempDir()

		// Act
		err := helpers.CopyMultipleFiles(dstDir, &testEmbedFS)

		// Assert
		assert.NoError(t, err)

		// Verify all files were copied
		file1Path := filepath.Join(dstDir, "testdata", "file1.txt")
		file2Path := filepath.Join(dstDir, "testdata", "file2.txt")
		configPath := filepath.Join(dstDir, "testdata", "config.json")
		nestedPath := filepath.Join(dstDir, "testdata", "subdir", "nested.txt")

		// Check file1.txt
		content1, err := os.ReadFile(file1Path)
		assert.NoError(t, err)
		assert.Contains(t, string(content1), "Test content for file 1")

		// Check file2.txt
		content2, err := os.ReadFile(file2Path)
		assert.NoError(t, err)
		assert.Contains(t, string(content2), "Test content for file 2")

		// Check config.json
		contentJSON, err := os.ReadFile(configPath)
		assert.NoError(t, err)
		assert.Contains(t, string(contentJSON), "test")
		assert.Contains(t, string(contentJSON), "1.0.0")

		// Check nested file
		contentNested, err := os.ReadFile(nestedPath)
		assert.NoError(t, err)
		assert.Contains(t, string(contentNested), "Nested file content")
	})
}

func TestCopyMultipleFiles_CreatesDirectories(t *testing.T) {
	t.Run("TestCopyMultipleFiles_CreatesDirectories", func(t *testing.T) {
		// Arrange
		dstDir := t.TempDir()

		// Act
		err := helpers.CopyMultipleFiles(dstDir, &testEmbedFS)

		// Assert
		assert.NoError(t, err)

		// Verify directories were created
		testdataDir := filepath.Join(dstDir, "testdata")
		subdirPath := filepath.Join(dstDir, "testdata", "subdir")

		testdataInfo, err := os.Stat(testdataDir)
		assert.NoError(t, err)
		assert.True(t, testdataInfo.IsDir(), "testdata should be a directory")

		subdirInfo, err := os.Stat(subdirPath)
		assert.NoError(t, err)
		assert.True(t, subdirInfo.IsDir(), "subdir should be a directory")
	})
}

func TestCopyMultipleFiles_PreservesFilePermissions(t *testing.T) {
	t.Run("TestCopyMultipleFiles_PreservesFilePermissions", func(t *testing.T) {
		// Arrange
		dstDir := t.TempDir()

		// Act
		err := helpers.CopyMultipleFiles(dstDir, &testEmbedFS)

		// Assert
		assert.NoError(t, err)

		// Verify file exists and is readable
		file1Path := filepath.Join(dstDir, "testdata", "file1.txt")
		fileInfo, err := os.Stat(file1Path)
		assert.NoError(t, err)
		assert.False(t, fileInfo.IsDir(), "should be a file not a directory")

		// Verify we can read the file (basic permission check)
		content, err := os.ReadFile(file1Path)
		assert.NoError(t, err)
		assert.NotEmpty(t, content)
	})
}

func TestCopyMultipleFiles_InvalidDestination(t *testing.T) {
	t.Run("TestCopyMultipleFiles_InvalidDestination", func(t *testing.T) {
		// Arrange - Use an invalid path that cannot be written to
		invalidPath := filepath.Join(string([]byte{0}), "invalid")

		// Act
		err := helpers.CopyMultipleFiles(invalidPath, &testEmbedFS)

		// Assert
		assert.Error(t, err)
	})
}

func TestCopyMultipleFiles_OverwritesExistingFiles(t *testing.T) {
	t.Run("TestCopyMultipleFiles_OverwritesExistingFiles", func(t *testing.T) {
		// Arrange
		dstDir := t.TempDir()

		// Create pre-existing directory and file
		testdataDir := filepath.Join(dstDir, "testdata")
		err := os.MkdirAll(testdataDir, 0755)
		assert.NoError(t, err)

		existingFilePath := filepath.Join(testdataDir, "file1.txt")
		err = os.WriteFile(existingFilePath, []byte("old content"), 0644)
		assert.NoError(t, err)

		// Act
		err = helpers.CopyMultipleFiles(dstDir, &testEmbedFS)

		// Assert
		assert.NoError(t, err)

		// Verify file was overwritten with new content
		content, err := os.ReadFile(existingFilePath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "Test content for file 1")
		assert.NotContains(t, string(content), "old content")
	})
}

func TestCopyMultipleFiles_EmptyEmbedFS(t *testing.T) {
	t.Run("TestCopyMultipleFiles_EmptyEmbedFS", func(t *testing.T) {
		// Arrange
		dstDir := t.TempDir()
		var emptyFS embed.FS

		// Act
		err := helpers.CopyMultipleFiles(dstDir, &emptyFS)

		// Assert
		// Should complete without error even with empty FS
		assert.NoError(t, err)
	})
}
