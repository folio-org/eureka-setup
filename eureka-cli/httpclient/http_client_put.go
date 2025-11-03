package httpclient

import (
	"net/http"
)

// HTTPClientPutManager defines the interface for HTTP PUT operations
type HTTPClientPutManager interface {
	PutReturnNoContent(url string, payload []byte, headers map[string]string) error
}

func (hc *HTTPClient) PutReturnNoContent(url string, payload []byte, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodPut, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}
