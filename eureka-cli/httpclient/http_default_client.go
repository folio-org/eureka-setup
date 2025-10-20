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
		Timeout: 30 * time.Minute,
		Transport: &LoggingRoundTripper{
			logger: slog.Default(),
			next: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Minute,
					KeepAlive: 90 * time.Second,
				}).DialContext,
				MaxIdleConns:           50,
				MaxIdleConnsPerHost:    10,
				IdleConnTimeout:        120 * time.Second,
				MaxResponseHeaderBytes: 16 << 20,
				WriteBufferSize:        64 << 10,
				ReadBufferSize:         64 << 10,
				ResponseHeaderTimeout:  5 * time.Minute,
				ExpectContinueTimeout:  10 * time.Second,
				DisableCompression:     false,
				ForceAttemptHTTP2:      false,
			},
		},
	}
}
