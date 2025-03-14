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
)

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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return ""
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "io.ReadAll error")
			panic(err)
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
			return ""
		}
	}

	return string(body)
}

func DoGetDecodeReturnInterface(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) interface{} {
	var respMap interface{}

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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
			return nil
		}
	}

	return respMap
}

func DoGetDecodeReturnMapStringInterface(commandName string, url string, enableDebug bool, panicOnError bool, headers map[string]string) map[string]interface{} {
	var respMap map[string]interface{}

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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
			return nil
		}
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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoPostReturnMapStringInteface(commandName string, url string, enableDebug bool, panicOnError bool, bodyBytes []byte, headers map[string]string) map[string]interface{} {
	var respMap map[string]interface{}

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
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		} else {
			LogWarn(commandName, enableDebug, fmt.Sprintf("json.NewDecoder warn - Cannot decode response from url: %s", url))
			return nil
		}
	}

	return respMap
}

func DoPostFormDataReturnMapStringInteface(commandName string, url string, enableDebug bool, formData url.Values, headers map[string]string) map[string]interface{} {
	var respMap map[string]interface{}

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
		resp.Body.Close()
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
		resp.Body.Close()
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
		} else {
			LogWarn(commandName, enableDebug, err.Error())
			return
		}
	}
	defer func() {
		CheckStatusCodes(commandName, panicOnError, resp)
		resp.Body.Close()
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
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}
