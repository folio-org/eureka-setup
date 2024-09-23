package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ####### HTTP ########

func DumpHttpBody(commandName string, enableDebug bool, bodyBytes []byte) {
	if !enableDebug {
		return
	}

	fmt.Println("###### Dumping HTTP Request Body ######")
	fmt.Println(string(bodyBytes))
	fmt.Println()
}

func DumpHttpFormData(commandName string, enableDebug bool, formData url.Values) {
	if !enableDebug {
		return
	}

	fmt.Println("###### Dumping HTTP Request Body ######")
	fmt.Println(formData)
	fmt.Println()
}

func DumpHttpRequest(commandName string, req *http.Request, enableDebug bool) {
	if !enableDebug {
		return
	}

	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error(commandName, "httputil.DumpRequest error", "")
		panic(err)
	}

	fmt.Println("###### Dumping HTTP Request ######")
	fmt.Println(string(reqBytes))
	fmt.Println()
}

func DumpHttpResponse(commandName string, resp *http.Response, enableDebug bool) {
	if !enableDebug {
		return
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error(commandName, "httputil.DumpResponse error", "")
		panic(err)
	}

	fmt.Println("###### Dumping HTTP Response ######")
	fmt.Println(string(respBytes))
	fmt.Println()
}

func CheckStatusCodes(commandName string, resp *http.Response) {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}

	LogErrorPanic(commandName, fmt.Sprintf("internal.CheckStatusCodes error - Unacceptable request status %d", resp.StatusCode))
}

func AddRequestHeaders(req *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		req.Header.Add(ContentTypeHeader, JsonContentType)
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

// ####### STRINGS ########

func TrimModuleName(name string) string {
	charIndex := strings.LastIndex(name, "-")
	if name[charIndex] == 45 {
		name = name[:charIndex]
	}

	return name
}

func TransformToEnvVar(name string) string {
	return EnvNameRegexp.ReplaceAllString(strings.ToUpper(name), "_")
}

// ######## LOG ########

func LogErrorPanic(commandName string, errorMessage string) {
	slog.Error(commandName, errorMessage, "")
	panic(errors.New(errorMessage))
}

func LogWarn(commandName string, enableDebug bool, errorMessage string) {
	if !enableDebug {
		return

	}
	slog.Warn(commandName, errorMessage, "")
}

// ######## SLICES ########

func ConvertMapKeysToSlice(inputMap map[string]any) []string {
	keys := make([]string, 0, len(inputMap))

	for key := range inputMap {
		keys = append(keys, key)
	}

	return keys
}
