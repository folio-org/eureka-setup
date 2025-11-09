package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
)

// HTTPClientGetManager defines the interface for HTTP GET operations
type HTTPClientGetManager interface {
	GetReturnStruct(url string, headers map[string]string, target any) error
	GetRetryReturnStruct(url string, headers map[string]string, target any) error
	GetReturnRawBytes(url string, headers map[string]string) ([]byte, error)
}

func (hc *HTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	return hc.getAndDecode(url, headers, false, target)
}

func (hc *HTTPClient) GetRetryReturnStruct(url string, headers map[string]string, target any) error {
	return hc.getAndDecode(url, headers, true, target)
}

func (hc *HTTPClient) GetReturnRawBytes(url string, headers map[string]string) ([]byte, error) {
	httpResponse, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
	if err != nil {
		return nil, err
	}
	defer CloseResponse(httpResponse)

	return io.ReadAll(httpResponse.Body)
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

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, target)
}
