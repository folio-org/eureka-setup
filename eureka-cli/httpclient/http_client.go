package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/hashicorp/go-retryablehttp"
)

type HTTPClientRunner interface {
	HTTPClientPinger
	HTTPClientGetManager
	HTTPClientPostManager
	HTTPClientPutManager
	HTTPClientDeleteManager
}

type HTTPClientPinger interface {
	Ping(url string) error
}

type HTTPClient struct {
	Action       *action.Action
	customClient *http.Client
	retryClient  *retryablehttp.Client
}

func New(action *action.Action, logger *slog.Logger) *HTTPClient {
	customClient := createCustomClient(constant.HTTPClientTimeout)
	return &HTTPClient{Action: action, customClient: customClient, retryClient: createRetryClient(logger, customClient)}
}

func (hc *HTTPClient) Ping(url string) error {
	resp, err := retryablehttp.Get(url)
	if err != nil {
		return fmt.Errorf("%s URL ping is unhealthy with error: %v", url, err)
	}
	defer CloseResponse(resp)

	slog.Info(hc.Action.Name, "text", "URL is accessible", "url", url)

	return nil
}

func (hc *HTTPClient) validateResponse(method, url string, resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	_ = helpers.DumpResponse(method, url, resp, true)

	return fmt.Errorf("unacceptable request status %d for URL: %s", resp.StatusCode, resp.Request.URL.String())
}

func (hc *HTTPClient) doRequest(method, url string, body []byte, headers map[string]string, useRetry bool) (*http.Response, error) {
	if body != nil {
		helpers.DumpRequestJSON(body)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	setRequestHeaders(req, headers)
	if err := helpers.DumpRequest(req); err != nil {
		return nil, err
	}

	var resp *http.Response
	if useRetry {
		retryReq, err := retryablehttp.FromRequest(req)
		if err != nil {
			return nil, err
		}
		resp, err = hc.retryClient.Do(retryReq)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err = hc.customClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if err := hc.validateResponse(method, url, resp); err != nil {
		CloseResponse(resp)
		return nil, err
	}

	if err := helpers.DumpResponse(method, url, resp, false); err != nil {
		CloseResponse(resp)
		return nil, err
	}

	return resp, nil
}

func setRequestHeaders(req *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		req.Header.Add(constant.ContentTypeHeader, constant.ApplicationJSON)
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

func CloseResponse(resp *http.Response) {
	_ = resp.Body.Close()
}
