package httpclient

import (
	"fmt"
	"net/http"

	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/hashicorp/go-retryablehttp"
)

// HTTPClientPinger defines the interface for HTTP ping operations
type HTTPClientPinger interface {
	PingRetry(url string) error
	Ping(url string) (int, error)
}

func (hc *HTTPClient) PingRetry(url string) error {
	statusCode, err := hc.doStatusCheck(url, true)
	if err != nil {
		return errors.PingFailed(url, err)
	}
	if statusCode != http.StatusOK {
		return errors.PingFailedWithStatus(url, statusCode)
	}

	return nil
}

func (hc *HTTPClient) Ping(url string) (int, error) {
	return hc.doStatusCheck(url, false)
}

func (hc *HTTPClient) doStatusCheck(url string, useRetry bool) (int, error) {
	httpRequest, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	var httpResponse *http.Response
	if useRetry {
		retryReq, err := retryablehttp.FromRequest(httpRequest)
		if err != nil {
			return 0, err
		}
		httpResponse, err = hc.retryClient.Do(retryReq)
		if err != nil {
			return 0, err
		}
	} else {
		httpResponse, err = hc.customClient.Do(httpRequest)
		if err != nil {
			return 0, err
		}
	}
	if httpResponse == nil {
		return 0, fmt.Errorf("received nil response from %s", url)
	}
	defer CloseResponse(httpResponse)

	return httpResponse.StatusCode, nil
}
