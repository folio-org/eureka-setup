package httpclient

import (
	"context"
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
	retryClient.CheckRetry = func(ctx context.Context, httpResponse *http.Response, err error) (bool, error) {
		// Use default retry policy for other errors
		shouldRetry, checkErr := retryablehttp.DefaultRetryPolicy(ctx, httpResponse, err)
		if shouldRetry {
			return true, checkErr
		}
		// Also retry on 429 Too Many Requests and 503 Service Unavailable
		if httpResponse != nil && (httpResponse.StatusCode == http.StatusTooManyRequests ||
			httpResponse.StatusCode == http.StatusServiceUnavailable) {
			return true, nil
		}

		return false, checkErr
	}
	retryClient.Logger = &LoggerAdapter{logger}

	return retryClient
}
