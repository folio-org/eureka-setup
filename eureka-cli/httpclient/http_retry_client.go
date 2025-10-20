package httpclient

import (
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/hashicorp/go-retryablehttp"
)

func createRetryClient() *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = constant.RetryHTTPClientRetryMax
	retryClient.RetryWaitMin = constant.RetryHTTPClientRetryWaitMin
	retryClient.RetryWaitMax = constant.RetryHTTPClientRetryWaitMax
	retryClient.Logger = slog.Default()

	return retryClient
}
