package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type HTTPClientGetManager interface {
	GetReturnResponse(url string, headers map[string]string) (*http.Response, error)
	GetRetryDecodeReturnAny(url string, headers map[string]string) (any, error)
	GetDecodeReturnMapStringAny(url string, headers map[string]string) (map[string]any, error)
	GetReturnStruct(url string, headers map[string]string, target any) error
}

func (hc *HTTPClient) GetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	return hc.doRequest(http.MethodGet, url, nil, headers, false)
}

func (hc *HTTPClient) GetRetryDecodeReturnAny(url string, headers map[string]string) (any, error) {
	resp, err := hc.doRequest(http.MethodGet, url, nil, headers, true)
	if err != nil {
		return nil, err
	}
	defer CloseResponse(resp)

	var respMap any
	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return respMap, nil
}

func (hc *HTTPClient) GetDecodeReturnMapStringAny(url string, headers map[string]string) (map[string]any, error) {
	resp, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
	if err != nil {
		return nil, err
	}
	defer CloseResponse(resp)

	var respMap map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return respMap, nil
}

func (hc *HTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	resp, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return json.NewDecoder(resp.Body).Decode(target)
}
