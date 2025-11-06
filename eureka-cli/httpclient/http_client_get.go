package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// HTTPClientGetManager defines the interface for HTTP GET operations
type HTTPClientGetManager interface {
	GetReturnResponse(url string, headers map[string]string) (*http.Response, error)
	GetReturnStruct(url string, headers map[string]string, target any) error
	GetRetryReturnStruct(url string, headers map[string]string, target any) error
}

func (hc *HTTPClient) GetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	return hc.doRequest(http.MethodGet, url, nil, headers, false)
}

func (hc *HTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	return hc.getAndDecode(url, headers, false, target)
}

func (hc *HTTPClient) GetRetryReturnStruct(url string, headers map[string]string, target any) error {
	return hc.getAndDecode(url, headers, true, target)
}

func (hc *HTTPClient) getAndDecode(url string, headers map[string]string, useRetry bool, target any) error {
	httpResponse, err := hc.doRequest(http.MethodGet, url, nil, headers, useRetry)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	if httpResponse.ContentLength == 0 {
		return nil
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}
