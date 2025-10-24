package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (hc *HTTPClient) PostReturnNoContent(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return nil
}

func (hc *HTTPClient) PostRetryReturnNoContent(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, true)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return nil
}

func (hc *HTTPClient) PostReturnMapStringAny(url string, b []byte, headers map[string]string) (map[string]any, error) {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, false)
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

func (hc *HTTPClient) PostFormDataReturnMapStringAny(url string, fd url.Values, headers map[string]string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(fd.Encode()))
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, headers)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer CloseResponse(resp)

	if err := hc.ValidateResponse(resp); err != nil {
		return nil, err
	}

	var respMap map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return respMap, nil
}
