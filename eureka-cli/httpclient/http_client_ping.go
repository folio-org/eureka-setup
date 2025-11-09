package httpclient

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/folio-org/eureka-cli/errors"
)

// HTTPClientPinger defines the interface for HTTP ping operations
type HTTPClientPinger interface {
	Ping(url string) error
	CheckStatus(url string) (int, error)
}

func (hc *HTTPClient) Ping(url string) error {
	statusCode, err := hc.doStatusCheck(url)
	if err != nil {
		return errors.PingFailed(url, err)
	}
	if statusCode != http.StatusOK {
		return errors.PingFailedWithStatus(url, statusCode)
	}
	slog.Info(hc.Action.Name, "text", "URL is accessible", "url", url)

	return nil
}

func (hc *HTTPClient) CheckStatus(url string) (int, error) {
	return hc.doStatusCheck(url)
}

func (hc *HTTPClient) doStatusCheck(url string) (int, error) {
	httpRequest, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	httpResponse, err := hc.pingClient.Do(httpRequest)
	if err != nil {
		return 0, err
	}
	if httpResponse == nil {
		return 0, fmt.Errorf("received nil response from %s", url)
	}
	defer CloseResponse(httpResponse)

	return httpResponse.StatusCode, nil
}
