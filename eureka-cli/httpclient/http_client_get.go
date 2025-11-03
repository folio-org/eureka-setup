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
	GetRetryDecodeReturnAny(url string, headers map[string]string) (any, error)
	GetReturnStruct(url string, headers map[string]string, target any) error
}

func (hc *HTTPClient) GetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	return hc.doRequest(http.MethodGet, url, nil, headers, false)
}

func (hc *HTTPClient) GetRetryDecodeReturnAny(url string, headers map[string]string) (any, error) {
	httpResponse, err := hc.doRequest(http.MethodGet, url, nil, headers, true)
	if err != nil {
		return nil, err
	}
	defer CloseResponse(httpResponse)

	var respMap any
	err = json.NewDecoder(httpResponse.Body).Decode(&respMap)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return respMap, nil
}

func (hc *HTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	httpResponse, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return json.NewDecoder(httpResponse.Body).Decode(target)
}
