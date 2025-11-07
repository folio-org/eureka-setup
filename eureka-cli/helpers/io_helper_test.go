package helpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestReadJsonFromFile_ValidJSON(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	jsonContent := `{"name": "test", "value": 123}`
	err := os.WriteFile(filePath, []byte(jsonContent), 0644)
	assert.NoError(t, err)

	var data map[string]any

	// Act
	err = helpers.ReadJsonFromFile("TestAction", filePath, &data)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test", data["name"])
	assert.Equal(t, float64(123), data["value"])
}

func TestReadJsonFromFile_FileNotFound(t *testing.T) {
	// Arrange
	var data map[string]any

	// Act
	err := helpers.ReadJsonFromFile("TestAction", "/nonexistent/file.json", &data)

	// Assert
	assert.Error(t, err)
}

func TestReadJsonFromFile_InvalidJSON(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(filePath, []byte("{invalid json"), 0644)
	assert.NoError(t, err)

	var data map[string]any

	// Act
	err = helpers.ReadJsonFromFile("TestAction", filePath, &data)

	// Assert
	assert.Error(t, err)
}

func TestWriteJsonToFile_ValidData(t *testing.T) {
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
	err = helpers.WriteJsonToFile(filePath, data)

	// Assert
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"name": "test"`)
	assert.Contains(t, string(content), `"value": 456`)
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
	err = helpers.CopySingleFile("TestAction", srcPath, dstPath)

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
	err := helpers.CopySingleFile("TestAction", "/nonexistent/source.txt", dstPath)

	// Assert
	assert.Error(t, err)
}

func TestCheckIsRegularFile_RegularFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "regular.txt")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)

	// Act
	err = helpers.CheckIsRegularFile("TestAction", filePath)

	// Assert
	assert.NoError(t, err)
}

func TestCheckIsRegularFile_NotFound(t *testing.T) {
	// Act
	err := helpers.CheckIsRegularFile("TestAction", "/nonexistent/file.txt")

	// Assert
	assert.Error(t, err)
}

func TestCheckIsRegularFile_Directory(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// Act
	err := helpers.CheckIsRegularFile("TestAction", tmpDir)

	// Assert
	assert.Error(t, err)
}

func TestGetCurrentWorkDirPath_Success(t *testing.T) {
	// Act
	result, err := helpers.GetCurrentWorkDirPath("TestAction")

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
	result, err := helpers.GetHomeMiscDir("TestAction")

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
