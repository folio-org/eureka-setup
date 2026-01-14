package httpclient

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
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
	lenientTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   constant.HTTPClientDialTimeout,
			KeepAlive: constant.HTTPClientKeepAlive,
		}).DialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          constant.HTTPClientMaxIdleConns,
		MaxIdleConnsPerHost:   constant.HTTPClientMaxIdleConnsPerHost,
		IdleConnTimeout:       constant.HTTPClientIdleConnTimeout,
		ResponseHeaderTimeout: constant.HTTPClientResponseHeaderTimeout,
		ExpectContinueTimeout: constant.HTTPClientExpectContinueTimeout,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &LoggingRoundTripper{logger: slog.Default(), next: lenientTransport},
	}
}

func createPingClient(timeout time.Duration) *http.Client {
	strictTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   constant.HTTPClientPingDialTimeout,
			KeepAlive: constant.HTTPClientPingKeepAlive,
		}).DialContext,
		DisableKeepAlives:     constant.HTTPClientPingDisableKeepAlives,
		MaxIdleConns:          constant.HTTPClientPingMaxIdleConns,
		MaxIdleConnsPerHost:   constant.HTTPClientPingMaxIdleConnsPerHost,
		IdleConnTimeout:       constant.HTTPClientPingIdleConnTimeout,
		ResponseHeaderTimeout: constant.HTTPClientPingResponseHeaderTimeout,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &LoggingRoundTripper{logger: slog.Default(), next: strictTransport},
	}
}
