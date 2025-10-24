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

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l.logger.Debug("HTTP request", "method", req.Method, "url", req.URL.String())

	start := time.Now()
	resp, err := l.next.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		l.logger.Error("HTTP request failed", "error", err, "duration", duration)

		return resp, err
	}

	l.logger.Debug("HTTP response", "method", req.Method, "status", resp.Status, "duration", duration)

	return resp, nil
}

func createCustomClient() *http.Client {
	return &http.Client{
		Timeout: constant.HTTPClientTimeout,
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
