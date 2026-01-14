package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
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

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, target)
}

func (hc *HTTPClient) PostFormDataReturnStruct(url string, formValues url.Values, headers map[string]string, target any) error {
	helpers.DumpRequestFormData(formValues)

	httpRequest, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formValues.Encode()))
	if err != nil {
		return err
	}

	setRequestHeaders(httpRequest, headers)
	if err := helpers.DumpRequest(httpRequest); err != nil {
		return err
	}

	httpResponse, err := hc.customClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer CloseResponse(httpResponse)

	if err := hc.validateResponse(url, http.MethodPost, httpResponse); err != nil {
		return err
	}
	if err := helpers.DumpResponse(http.MethodPost, url, httpResponse, false); err != nil {
		return err
	}
	if httpResponse.ContentLength == 0 {
		return nil
	}

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, target)
}
