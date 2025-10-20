package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

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

// ####### GET ########

func (hc *HTTPClient) DoGetReturnResponse(url string, panicOnError bool, headers map[string]string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}

	helpers.DumpResponse(hc.Action, resp, false)

	return resp
}

func (hc *HTTPClient) DoGetDecodeReturnString(url string, panicOnError bool, headers map[string]string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return ""
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return ""
	}

	return string(body)
}

func (hc *HTTPClient) DoGetDecodeReturnAny(url string, panicOnError bool, headers map[string]string) any {
	var respMap any

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}

	return respMap
}

func (hc *HTTPClient) DoGetDecodeReturnMapStringAny(url string, panicOnError bool, headers map[string]string) map[string]any {
	var respMap map[string]any

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}

	return respMap
}

// ####### POST ########

func (hc *HTTPClient) DoPostReturnNoContent(url string, panicOnError bool, bodyBytes []byte, headers map[string]string) {
	helpers.DumpJSONBodyRequestBody(bodyBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)
}

func (hc *HTTPClient) DoRetryPostReturnNoContent(url string, panicOnError bool, bodyBytes []byte, headers map[string]string) {
	helpers.DumpJSONBodyRequestBody(bodyBytes)

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req.Request, headers)
	helpers.DumpRequest(hc.Action, req.Request)

	resp, err := hc.retryClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)
}

func (hc *HTTPClient) DoPostReturnMapStringAny(url string, panicOnError bool, bodyBytes []byte, headers map[string]string) map[string]any {
	var respMap map[string]any

	helpers.DumpJSONBodyRequestBody(bodyBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return nil
	}

	return respMap
}

func (hc *HTTPClient) DoPostFormDataReturnMapStringAny(url string, formData url.Values, headers map[string]string) map[string]any {
	var respMap map[string]any

	helpers.DumpFormDataRequestBody(formData)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formData.Encode()))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}
	defer func() {
		hc.checkStatusCodes(true, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	return respMap
}

// ####### PUT ########

func (hc *HTTPClient) DoPutReturnNoContent(url string, bodyBytes []byte, headers map[string]string) {
	helpers.DumpJSONBodyRequestBody(bodyBytes)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}
	defer func() {
		hc.checkStatusCodes(true, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)
}

// ####### DELETE ########

func (hc *HTTPClient) DoDelete(url string, panicOnError bool, headers map[string]string) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(hc.Action.Name, "error", err)
			panic(err)
		}

		helpers.LogDebug(hc.Action, err)

		return
	}
	defer func() {
		hc.checkStatusCodes(panicOnError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)
}

func (hc *HTTPClient) DoDeleteWithBody(url string, bodyBytes []byte, ignoreError bool, headers map[string]string) {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}

	addRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		slog.Error(hc.Action.Name, "error", err)
		panic(err)
	}
	defer func() {
		hc.checkStatusCodes(!ignoreError, resp)
		_ = resp.Body.Close()
	}()

	helpers.DumpResponse(hc.Action, resp, false)
}

func (hc *HTTPClient) checkStatusCodes(panicOnError bool, resp *http.Response) {
	if !panicOnError || (resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices) {
		return
	}

	helpers.DumpResponse(hc.Action, resp, true)
	helpers.LogErrorPanic(hc.Action, fmt.Errorf("unacceptable request status %d for URL: %s", resp.StatusCode, resp.Request.URL.String()))
}

func addRequestHeaders(req *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		req.Header.Add(constant.ContentTypeHeader, constant.JsonContentType)
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}
