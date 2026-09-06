package httpclient

import (
	"log/slog"
	"net/http"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/hashicorp/go-retryablehttp"
)

// LoggerAdapter adapts slog.Logger to retryablehttp.LeveledLogger interface
type LoggerAdapter struct {
	logger *slog.Logger
}

func (l *LoggerAdapter) Error(msg string, keysAndValues ...any) {
	l.logger.Error(msg, keysAndValues...)
}

func (l *LoggerAdapter) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *LoggerAdapter) Debug(msg string, keysAndValues ...any) {
	l.logger.Debug(msg, keysAndValues...)
}

func (l *LoggerAdapter) Warn(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, keysAndValues...)
}

func createRetryClient(logger *slog.Logger, customClient *http.Client) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = customClient
	retryClient.RetryMax = constant.RetryHTTPClientRetryMax
	retryClient.RetryWaitMin = constant.RetryHTTPClientRetryWaitMin
	retryClient.RetryWaitMax = constant.RetryHTTPClientRetryWaitMax
	retryClient.Logger = &LoggerAdapter{logger}

	return retryClient
}
