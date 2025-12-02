package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
)

// HTTPClientDeleteManager defines the interface for HTTP DELETE operations
type HTTPClientDeleteManager interface {
	Delete(url string, headers map[string]string) error
	DeleteReturnStruct(url string, headers map[string]string, target any) error
	DeleteWithPayloadReturnStruct(url string, payload []byte, headers map[string]string, target any) error
}

func (hc *HTTPClient) Delete(url string, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodDelete, url, nil, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}

func (hc *HTTPClient) DeleteReturnStruct(url string, headers map[string]string, target any) error {
	httpResponse, err := hc.doRequest(http.MethodDelete, url, nil, headers, false)
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

func (hc *HTTPClient) DeleteWithPayloadReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	httpResponse, err := hc.doRequest(http.MethodDelete, url, payload, headers, false)
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
