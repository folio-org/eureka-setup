package httpclient

import (
	"bytes"
	"log/slog"
	"testing"
)

// TestLoggerAdapter tests the LoggerAdapter methods
func TestLoggerAdapter_Error(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := &LoggerAdapter{logger: logger}

	// Act
	adapter.Error("test error message", "key", "value")

	// Assert
	output := buf.String()
	if output == "" {
		t.Error("Expected error log output, got empty string")
	}
	if !contains(output, "test error message") {
		t.Errorf("Expected output to contain 'test error message', got: %s", output)
	}
}

func TestLoggerAdapter_Info(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	adapter := &LoggerAdapter{logger: logger}

	// Act
	adapter.Info("test info message", "action", "retry")

	// Assert
	output := buf.String()
	if output == "" {
		t.Error("Expected info log output, got empty string")
	}
	if !contains(output, "test info message") {
		t.Errorf("Expected output to contain 'test info message', got: %s", output)
	}
}

func TestLoggerAdapter_Debug(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := &LoggerAdapter{logger: logger}

	// Act
	adapter.Debug("test debug message", "detail", "trace")

	// Assert
	output := buf.String()
	if output == "" {
		t.Error("Expected debug log output, got empty string")
	}
	if !contains(output, "test debug message") {
		t.Errorf("Expected output to contain 'test debug message', got: %s", output)
	}
}

func TestLoggerAdapter_Warn(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	adapter := &LoggerAdapter{logger: logger}

	// Act
	adapter.Warn("test warn message", "reason", "timeout")

	// Assert
	output := buf.String()
	if output == "" {
		t.Error("Expected warn log output, got empty string")
	}
	if !contains(output, "test warn message") {
		t.Errorf("Expected output to contain 'test warn message', got: %s", output)
	}
}

func TestLoggerAdapter_WithKeyValuePairs(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	adapter := &LoggerAdapter{logger: logger}

	// Act
	adapter.Info("operation", "key1", "value1", "key2", "value2")

	// Assert
	output := buf.String()
	if output == "" {
		t.Error("Expected log output with key-value pairs")
	}
	if !contains(output, "operation") {
		t.Errorf("Expected output to contain 'operation', got: %s", output)
	}
}

func TestLoggerAdapter_AllLevels(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := &LoggerAdapter{logger: logger}

	// Act - Test all log levels
	adapter.Debug("debug log")
	adapter.Info("info log")
	adapter.Warn("warn log")
	adapter.Error("error log")

	// Assert
	output := buf.String()
	if !contains(output, "debug log") {
		t.Error("Expected debug log in output")
	}
	if !contains(output, "info log") {
		t.Error("Expected info log in output")
	}
	if !contains(output, "warn log") {
		t.Error("Expected warn log in output")
	}
	if !contains(output, "error log") {
		t.Error("Expected error log in output")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// Integration tests for retry behavior

func TestCreateRetryClient_Configuration(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	customClient := createCustomClient(5)

	// Act
	retryClient := createRetryClient(logger, customClient)

	// Assert
	if retryClient == nil {
		t.Fatal("Expected non-nil retry client")
	}
	if retryClient.HTTPClient != customClient {
		t.Error("Expected custom HTTP client to be set")
	}
	if retryClient.Logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestCreateRetryClient_LoggerAdapterSet(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	customClient := createCustomClient(5)

	// Act
	retryClient := createRetryClient(logger, customClient)

	// Assert
	adapter, ok := retryClient.Logger.(*LoggerAdapter)
	if !ok {
		t.Fatal("Expected logger to be LoggerAdapter type")
	}
	if adapter.logger != logger {
		t.Error("Expected LoggerAdapter to wrap the provided logger")
	}
}
