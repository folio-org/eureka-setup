package httpclient

import (
	"fmt"
	"net/http"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/hashicorp/go-retryablehttp"
)

type HTTPClient struct {
	Action       *action.Action
	customClient *http.Client
	retryClient  *retryablehttp.Client
}

func New(action *action.Action) *HTTPClient {
	return &HTTPClient{
		Action:       action,
		customClient: createCustomClient(),
		retryClient:  createRetryClient(),
	}
}

func SetRequestHeaders(req *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		req.Header.Add(constant.ContentTypeHeader, constant.ApplicationJSON)
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

func (hc *HTTPClient) ValidateResponse(resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	helpers.DumpResponse(hc.Action, resp, true)

	return fmt.Errorf("unacceptable request status %d for URL: %s", resp.StatusCode, resp.Request.URL.String())
}
