package httpclient

import (
	"net/http"
)

// HTTPClientDeleteManager defines the interface for HTTP DELETE operations
type HTTPClientDeleteManager interface {
	Delete(url string, headers map[string]string) error
	DeleteWithBody(url string, payload []byte, headers map[string]string) error
}

func (hc *HTTPClient) Delete(url string, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodDelete, url, nil, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}

func (hc *HTTPClient) DeleteWithBody(url string, payload []byte, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodDelete, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}
