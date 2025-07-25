package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	DefaultRetryMax int = 10

	DefaultRetryWaitMax time.Duration = 5 * time.Second
)

func createRetryableClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = DefaultRetryMax
	client.RetryWaitMax = DefaultRetryWaitMax
	client.Logger = nil

	return client
}

// ####### GET ########

func DoGetReturnResponse(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return nil
	}

	DumpHttpResponse(commandName, resp, enableDebug)

	return resp
}

func DoGetDecodeReturnString(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return ""
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "io.ReadAll error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
		return ""
	}

	return string(body)
}

func DoGetDecodeReturnAny(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) any {
	var respMap any

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return nil
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
		return nil
	}

	return respMap
}

func DoGetDecodeReturnMapStringAny(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) map[string]any {
	var respMap map[string]any

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return nil
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
		return nil
	}

	return respMap
}

// ####### POST ########

func DoPostReturnNoContent(commandName string, url string, enableDebug bool, panicOnError bool, bodyBytes []byte, headers map[string]string) {
	DumpHttpBody(commandName, enableDebug, bodyBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoRetryablePostReturnNoContent(commandName string, url string, enableDebug bool, panicOnError bool, bodyBytes []byte, headers map[string]string) {
	DumpHttpBody(commandName, enableDebug, bodyBytes)

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req.Request, headers)
	DumpHttpRequest(commandName, req.Request, enableDebug)

	resp, err := createRetryableClient().Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoPostReturnMapStringAny(commandName string, url string, enableDebug bool, panicOnError bool, bodyBytes []byte, headers map[string]string) map[string]any {
	var respMap map[string]any

	DumpHttpBody(commandName, enableDebug, bodyBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
		return nil
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
		return nil
	}

	return respMap
}

func DoPostFormDataReturnMapStringAny(commandName string, url string, enableDebug bool, formData url.Values, headers map[string]string) map[string]any {
	var respMap map[string]any

	DumpHttpFormData(commandName, enableDebug, formData)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formData.Encode()))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, true, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
		panic(err)
	}

	return respMap
}

// ####### PUT ########

func DoPutReturnNoContent(commandName string, url string, enableDebug bool, bodyBytes []byte, headers map[string]string) {
	DumpHttpBody(commandName, enableDebug, bodyBytes)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, true, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

// ####### DELETE ########

func DoDelete(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
			panic(err)
		}

		LogWarn(commandName, enableDebug, err.Error())
		return
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoDeleteWithBody(commandName string, url string, enableDebug bool, bodyBytes []byte, ignoreError bool, headers map[string]string) {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.NewRequest error")
		panic(err)
	}

	AddRequestHeaders(req, headers)
	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "http.DefaultClient.Do error")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, !ignoreError, resp)
		_ = resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}
