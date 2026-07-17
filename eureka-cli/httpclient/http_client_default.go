package httpclient

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
)

type LoggingRoundTripper struct {
	logger *slog.Logger
	next   http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(httpRequest *http.Request) (*http.Response, error) {
	var targetPristine string
	var targetHyphenated string

	// --- Request Mutation: Translate route hyphen back to pristine '+' for lookup ---
	if httpRequest.URL != nil && strings.Contains(httpRequest.URL.Path, "/_/proxy/modules/mod-") {
		if idx := strings.Index(httpRequest.URL.Path, "-SNAPSHOT."); idx != -1 {
			rem := httpRequest.URL.Path[idx+10:]
			if hIdx := strings.Index(rem, "-"); hIdx != -1 {
				segment := rem[:hIdx]
				isDigits := len(segment) > 0
				for _, ch := range segment {
					if ch < '0' || ch > '9' {
						isDigits = false
						break
					}
				}
				if isDigits && hIdx+1 < len(rem) && rem[hIdx+1] >= '0' && rem[hIdx+1] <= '9' {
					exactIdx := idx + 10 + hIdx
					targetHyphenated = httpRequest.URL.Path[idx:]

					// Mutate target path for outbound wire handshake
					httpRequest.URL.Path = httpRequest.URL.Path[:exactIdx] + "+" + httpRequest.URL.Path[exactIdx+1:]
					httpRequest.URL.RawPath = ""

					targetPristine = httpRequest.URL.Path[idx:]
				}
			}
		}
	}

	l.logger.Debug("HTTP request", "method", httpRequest.Method, "url", httpRequest.URL.String())
	start := time.Now()
	httpResponse, err := l.next.RoundTrip(httpRequest)
	duration := time.Since(start)

	if err != nil {
		l.logger.Debug("HTTP request failed", "error", err, "duration", duration)
		return httpResponse, err
	}

	// --- Response Normalization: Enforce safe routing hyphen structure inside the payload ---
	if httpResponse.StatusCode == http.StatusOK && targetPristine != "" && httpResponse.Body != nil {
		bodyBytes, readErr := io.ReadAll(httpResponse.Body)
		httpResponse.Body.Close()
		if readErr == nil {
			bodyStr := string(bodyBytes)
			// Align descriptor interior ID with module manifest declarations
			bodyStr = strings.ReplaceAll(bodyStr, targetPristine, targetHyphenated)

			httpResponse.Body = io.NopCloser(strings.NewReader(bodyStr))
			httpResponse.Header.Del("Content-Length") // Let net/http re-calculate frame lengths
		}
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