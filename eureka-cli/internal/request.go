package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// ####### GET ########

func DoGetReturnResponse(commandName string, url string, enableDebug bool, panicOnError bool) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	req.Header.Set("Content-Type", JsonContentType)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, "http.DefaultClient.Do error", "")
			panic(err)
		} else {
			LogWarn(commandName, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
	}

	DumpHttpResponse(commandName, resp, enableDebug)

	return resp
}

func DoGetDecodeReturnInterface(commandName string, url string, enableDebug bool) interface{} {
	var respMap interface{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	req.Header.Set("Content-Type", JsonContentType)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, "http.DefaultClient.Do error", "")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, "json.NewDecoder error", "")
		panic(err)
	}

	return respMap
}

func DoGetDecodeReturnMapStringInteface(commandName string, url string, enableDebug bool, panicOnError bool) map[string]interface{} {
	var respMap map[string]interface{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	req.Header.Set("Content-Type", JsonContentType)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if panicOnError {
			slog.Error(commandName, "http.DefaultClient.Do error", "")
			panic(err)
		} else {
			LogWarn(commandName, fmt.Sprintf("http.DefaultClient.Do warn - Endpoint is unreachable: %s", url))
			return nil
		}
	}
	defer func() {
		CheckStatusCodes(commandName, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, "json.NewDecoder error", "")
		panic(err)
	}

	return respMap
}

// ####### POST ########

func DoPostReturnNoContent(commandName string, url string, enableDebug bool, bodyBytes []byte, headers map[string]string) {
	DumpHttpBody(commandName, enableDebug, bodyBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, "http.DefaultClient.Do error", "")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoPostReturnMapStringInteface(commandName string, url string, enableDebug bool, formData url.Values, headers map[string]string) map[string]interface{} {
	var respMap map[string]interface{}

	DumpHttpFormData(commandName, enableDebug, formData)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formData.Encode()))
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, "http.DefaultClient.Do error", "")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, "json.NewDecoder error", "")
		panic(err)
	}

	return respMap
}

// ####### DELETE ########

func DoDelete(commandName string, url string, enableDebug bool) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, "http.DefaultClient.Do error", "")
		panic(err)
	}
	defer func() {
		CheckStatusCodes(commandName, resp)
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}

func DoDeleteWithBody(commandName string, url string, enableDebug bool, bodyBytes []byte, ignoreError bool) {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Error(commandName, "http.NewRequest error", "")
		panic(err)
	}

	req.Header.Add("Content-Type", JsonContentType)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, "http.DefaultClient.Do error", "")
		panic(err)
	}
	defer func() {
		if !ignoreError {
			CheckStatusCodes(commandName, resp)
		}
		resp.Body.Close()
	}()

	DumpHttpResponse(commandName, resp, enableDebug)
}
