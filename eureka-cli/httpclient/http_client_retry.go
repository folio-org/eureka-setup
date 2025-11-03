package httpclient

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/hashicorp/go-retryablehttp"
)

func createRetryClient(logger *slog.Logger, customClient *http.Client) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = customClient
	retryClient.RetryMax = constant.RetryHTTPClientRetryMax
	retryClient.RetryWaitMin = constant.RetryHTTPClientRetryWaitMin
	retryClient.RetryWaitMax = constant.RetryHTTPClientRetryWaitMax
	retryClient.CheckRetry = func(ctx context.Context, httpResponse *http.Response, err error) (bool, error) {
		shouldRetry, checkErr := retryablehttp.DefaultRetryPolicy(ctx, httpResponse, err)
		if shouldRetry {
			return true, checkErr
		}
		if httpResponse != nil && (httpResponse.StatusCode == http.StatusTooManyRequests ||
			httpResponse.StatusCode == http.StatusServiceUnavailable) {
			return true, nil
		}

		return false, checkErr
	}
	retryClient.Logger = logger

	return retryClient
}
