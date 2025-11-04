package httpclient

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/folio-org/eureka-cli/constant"
)

type LoggingRoundTripper struct {
	logger *slog.Logger
	next   http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(httpRequest *http.Request) (*http.Response, error) {
	l.logger.Debug("HTTP request", "method", httpRequest.Method, "url", httpRequest.URL.String())
	start := time.Now()
	httpResponse, err := l.next.RoundTrip(httpRequest)
	duration := time.Since(start)
	if err != nil {
		l.logger.Debug("HTTP request failed", "error", err, "duration", duration)

		return httpResponse, err
	}
	l.logger.Debug("HTTP response", "method", httpRequest.Method, "status", httpResponse.Status, "duration", duration)

	return httpResponse, nil
}

func createCustomClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &LoggingRoundTripper{
			logger: slog.Default(),
			next: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   constant.HTTPClientDialTimeout,
					KeepAlive: constant.HTTPClientKeepAlive,
				}).DialContext,
				MaxIdleConns:           constant.HTTPClientMaxIdleConns,
				MaxIdleConnsPerHost:    constant.HTTPClientMaxIdleConnsPerHost,
				IdleConnTimeout:        constant.HTTPClientIdleConnTimeout,
				MaxResponseHeaderBytes: constant.HTTPClientMaxResponseHeaderBytes,
				WriteBufferSize:        constant.HTTPClientWriteBufferSize,
				ReadBufferSize:         constant.HTTPClientReadBufferSize,
				ResponseHeaderTimeout:  constant.HTTPClientResponseHeaderTimeout,
				ExpectContinueTimeout:  constant.HTTPClientExpectContinueTimeout,
				DisableCompression:     constant.HTTPClientDisableCompression,
				ForceAttemptHTTP2:      constant.HTTPClientForceAttemptHTTP2,
			},
		},
	}
}
