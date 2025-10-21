package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (hc *HTTPClient) GetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	return hc.doRequest(http.MethodGet, url, nil, headers, false)
}

func (hc *HTTPClient) GetDecodeReturnString(url string, headers map[string]string) (string, error) {
	resp, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
	if err != nil {
		return "", err
	}
	defer CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (hc *HTTPClient) GetDecodeReturnAny(url string, headers map[string]string) (any, error) {
	resp, err := hc.doRequest(http.MethodGet, url, nil, headers, false)
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
