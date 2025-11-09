package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
)

// HTTPClientPutManager defines the interface for HTTP PUT operations
type HTTPClientPutManager interface {
	PutReturnNoContent(url string, payload []byte, headers map[string]string) error
	PutReturnStruct(url string, payload []byte, headers map[string]string, target any) error
}

func (hc *HTTPClient) PutReturnNoContent(url string, payload []byte, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodPut, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}

func (hc *HTTPClient) PutReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	httpResponse, err := hc.doRequest(http.MethodPut, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, target)
}
