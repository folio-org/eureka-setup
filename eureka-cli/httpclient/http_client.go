package httpclient

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/hashicorp/go-retryablehttp"
)

// HTTPClientRunner defines the interface for HTTP client operations
type HTTPClientRunner interface {
	HTTPClientPinger
	HTTPClientGetManager
	HTTPClientPostManager
	HTTPClientPutManager
	HTTPClientDeleteManager
}

// HTTPClient provides functionality for HTTP client operations with retry logic
type HTTPClient struct {
	Action       *action.Action
	customClient *http.Client
	retryClient  *retryablehttp.Client
	pingClient   *retryablehttp.Client
}

// New creates a new HTTPClient instance
func New(action *action.Action, logger *slog.Logger) *HTTPClient {
	customClient := createCustomClient(constant.HTTPClientTimeout)
	pingClient := createPingClient(constant.HTTPClientPingTimeout)
	return &HTTPClient{
		Action:       action,
		customClient: customClient,
		retryClient:  createRetryClient(logger, customClient),
		pingClient:   createRetryClient(logger, pingClient),
	}
}

func (hc *HTTPClient) doRequest(method, url string, payload []byte, headers map[string]string, useRetry bool) (*http.Response, error) {
	if payload != nil {
		helpers.DumpRequestJSON(payload)
	}

	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}

	httpRequest, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	setRequestHeaders(httpRequest, headers)
	if err := helpers.DumpRequest(httpRequest); err != nil {
		return nil, err
	}

	var httpResponse *http.Response
	if useRetry {
		retryReq, err := retryablehttp.FromRequest(httpRequest)
		if err != nil {
			return nil, err
		}
		httpResponse, err = hc.retryClient.Do(retryReq)
		if err != nil {
			return nil, err
		}
	} else {
		httpResponse, err = hc.customClient.Do(httpRequest)
		if err != nil {
			return nil, err
		}
	}
	if err := hc.validateResponse(method, url, httpResponse); err != nil {
		CloseResponse(httpResponse)
		return nil, err
	}
	if err := helpers.DumpResponse(method, url, httpResponse, false); err != nil {
		CloseResponse(httpResponse)
		return nil, err
	}

	return httpResponse, nil
}

func setRequestHeaders(httpRequest *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		httpRequest.Header.Add(constant.ContentTypeHeader, constant.ApplicationJSON)
		return
	}
	for key, value := range headers {
		httpRequest.Header.Add(key, value)
	}
}

func (hc *HTTPClient) validateResponse(method, url string, httpResponse *http.Response) error {
	if httpResponse.StatusCode >= http.StatusOK && httpResponse.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	_ = helpers.DumpResponse(method, url, httpResponse, true)

	return errors.RequestFailed(httpResponse.StatusCode, httpResponse.Request.Method, httpResponse.Request.URL.String())
}

func CloseResponse(httpResponse *http.Response) {
	if httpResponse != nil && httpResponse.Body != nil {
		// Drain any remaining data to enable connection reuse
		_, _ = io.Copy(io.Discard, httpResponse.Body)
		_ = httpResponse.Body.Close()
	}
}
