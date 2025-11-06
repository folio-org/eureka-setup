package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPClientPostManager defines the interface for HTTP POST operations
type HTTPClientPostManager interface {
	PostReturnNoContent(url string, payload []byte, headers map[string]string) error
	PostRetryReturnNoContent(url string, payload []byte, headers map[string]string) error
	PostReturnStruct(url string, payload []byte, headers map[string]string, target any) error
	PostFormDataReturnStruct(url string, formValues url.Values, headers map[string]string, target any) error
}

func (hc *HTTPClient) PostReturnNoContent(url string, payload []byte, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodPost, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}

func (hc *HTTPClient) PostRetryReturnNoContent(url string, payload []byte, headers map[string]string) error {
	httpResponse, err := hc.doRequest(http.MethodPost, url, payload, headers, true)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	return nil
}

func (hc *HTTPClient) PostReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	httpResponse, err := hc.doRequest(http.MethodPost, url, payload, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	if httpResponse.ContentLength == 0 {
		return nil
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}

func (hc *HTTPClient) PostFormDataReturnStruct(url string, formValues url.Values, headers map[string]string, target any) error {
	httpRequest, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formValues.Encode()))
	if err != nil {
		return err
	}

	setRequestHeaders(httpRequest, headers)
	httpResponse, err := hc.customClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	if err := hc.validateResponse(url, http.MethodPost, httpResponse); err != nil {
		return err
	}

	if httpResponse.ContentLength == 0 {
		return nil
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}
