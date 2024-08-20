package internal

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

// ####### GET ########

func DoGetReturnResponse(commandName string, url string, enableDebug bool) *http.Response {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.NewRequest error")
		panic(err)
	}

	req.Header.Set("Content-Type", ApplicationJson)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.DefaultClient.Do error")
		panic(err)
	}

	DumpHttpResponse(commandName, resp, enableDebug)

	return resp
}

func DoGetDecodeReturnInterface(commandName string, url string, enableDebug bool) interface{} {
	var respMap interface{}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.NewRequest error")
		panic(err)
	}

	req.Header.Set("Content-Type", ApplicationJson)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, MessageKey, "json.NewDecoder error")
		panic(err)
	}

	return respMap
}

func DoGetDecodeReturnMapStringInteface(commandName string, url string, enableDebug bool) map[string]interface{} {
	var respMap map[string]interface{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.NewRequest error")
		panic(err)
	}

	req.Header.Set("Content-Type", ApplicationJson)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	DumpHttpResponse(commandName, resp, enableDebug)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		slog.Error(commandName, MessageKey, "json.NewDecoder error")
		panic(err)
	}

	return respMap
}

// ####### POST ########

func DoPostNoContent(commandName string, url string, enableDebug bool, payloadBytes []byte) {
	DumpHttpBody(commandName, enableDebug, payloadBytes)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error(commandName, MessageKey, "http.NewRequest error")
		panic(err)
	}

	req.Header.Add("Content-Type", ApplicationJson)

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	DumpHttpResponse(commandName, resp, enableDebug)
}

// ####### DELETE ########

func DoDelete(commandName string, url string, enableDebug bool) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.NewRequest error")
		panic(err)
	}

	DumpHttpRequest(commandName, req, enableDebug)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, MessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	DumpHttpResponse(commandName, resp, enableDebug)
}
