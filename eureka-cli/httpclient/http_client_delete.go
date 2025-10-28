package httpclient

import (
	"net/http"
)

type HTTPClientDeleteManager interface {
	Delete(url string, headers map[string]string) error
	DeleteWithBody(url string, b []byte, headers map[string]string) error
}

func (hc *HTTPClient) Delete(url string, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodDelete, url, nil, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return nil
}

func (hc *HTTPClient) DeleteWithBody(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodDelete, url, b, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return nil
}
