package httpclient

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func (hc *HTTPClient) PostReturnNoContent(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, false)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}

func (hc *HTTPClient) RetryPostReturnNoContent(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, true)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}

func (hc *HTTPClient) PostReturnMapStringAny(url string, b []byte, headers map[string]string) (map[string]any, error) {
	resp, err := hc.doRequest(http.MethodPost, url, b, headers, false)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	var respMap map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respMap); err != nil {
		return nil, err
	}

	return respMap, nil
}

func (hc *HTTPClient) PostFormDataReturnMapStringAny(url string, fd url.Values, headers map[string]string) (map[string]any, error) {
	// For form data, we'll use the old pattern for now since it's special
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(fd.Encode()))
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, headers)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if err := hc.ValidateResponse(resp); err != nil {
		return nil, err
	}

	var respMap map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respMap); err != nil {
		return nil, err
	}

	return respMap, nil
}
