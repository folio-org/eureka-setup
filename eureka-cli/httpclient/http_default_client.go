package httpclient

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

type LoggingRoundTripper struct {
	logger *slog.Logger
	next   http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l.logger.Info("HTTP request", "method", req.Method, "url", req.URL.String())

	start := time.Now()
	resp, err := l.next.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		l.logger.Error("HTTP request failed", "error", err, "duration", duration)
		return resp, err
	}

	l.logger.Info("HTTP response", "status", resp.Status, "duration", duration)

	return resp, nil
}

func createCustomClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &LoggingRoundTripper{
			logger: slog.Default(),
			next: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				IdleConnTimeout:       90 * time.Second,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
			},
		},
	}
}
