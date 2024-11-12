package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
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
		slog.Error(commandName, GetFuncName(), "httputil.DumpRequest error")
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
		slog.Error(commandName, GetFuncName(), "httputil.DumpResponse error")
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
	slog.Error(commandName, GetFuncName(), errorMessage)
	panic(errors.New(errorMessage))
}

func LogWarn(commandName string, enableDebug bool, errorMessage string) {
	if !enableDebug {
		return

	}
	slog.Warn(commandName, GetFuncName(), errorMessage)
}

// ######## SLICES ########

func ConvertMapKeysToSlice(inputMap map[string]any) []string {
	keys := make([]string, 0, len(inputMap))

	for key := range inputMap {
		keys = append(keys, key)
	}

	return keys
}

// ######## JSON ########

func ReadJsonFromFile(commandName string, filePath string, data interface{}) {
	jsonFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.Open error")
		panic(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	for {
		if err := decoder.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			slog.Error(commandName, GetFuncName(), "decoder.Decode error")
			panic(err)
		}
	}
}

func WriteJsonToFile(commandName string, filePath string, packageJson interface{}) {
	jsonFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.Open error")
		panic(err)
	}
	defer jsonFile.Close()

	writer := bufio.NewWriter(jsonFile)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err = encoder.Encode(packageJson); err != nil {
		slog.Error(commandName, GetFuncName(), "encoder.Encode error")
		panic(err)
	}

	writer.Flush()
}

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)

	return runtime.FuncForPC(pc).Name()
}
