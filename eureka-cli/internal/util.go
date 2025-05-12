package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
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

func CheckStatusCodes(commandName string, panicOnError bool, resp *http.Response) {
	if !panicOnError || resp.StatusCode >= 200 && resp.StatusCode < 300 {
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

// ######## LOG ########

func LogErrorPanic(commandName string, errorMessage string) {
	slog.Error(commandName, GetFuncName(), errorMessage)
	panic(errors.New(errorMessage))
}

func LogErrorPrintStderrPanic(commandName string, errorMessage string, stackTrace string) {
	slog.Error(commandName, GetFuncName(), errorMessage)
	fmt.Println("Stderr: ", stackTrace)
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

// ######## IO ########

func ReadJsonFromFile(commandName string, filePath string, data any) {
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

func WriteJsonToFile(commandName string, filePath string, packageJson any) {
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

func CreateFile(commandName string, fileName string) *os.File {
	filePointer, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.OpenFile error")
		panic(err)
	}

	return filePointer
}

func CheckIsRegularFile(commandName string, fileName string) {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.Stat error")
		panic(err)
	}

	if !fileStat.Mode().IsRegular() {
		LogErrorPanic(commandName, "fileStat.Mode().IsRegular error")
	}
}

func CopySingleFile(commandName string, srcPath string, dstPath string) {
	CheckIsRegularFile(commandName, srcPath)

	src, err := os.Open(srcPath)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.Open error")
		panic(err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.Create error")
		panic(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "io.Copy error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Copied a single file from %s to %s", filepath.FromSlash(srcPath), filepath.FromSlash(dstPath)))
}

// ######## Runtime ########

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)

	return runtime.FuncForPC(pc).Name()
}

// ######## Net ########

func GetFreePortFromRange(commandName string, portStart, portEnd int, excludeReservedPorts []int) int {
	for port := portStart + 1; port <= portEnd; port++ {
		if !slices.Contains(excludeReservedPorts, port) && IsPortFree(commandName, portStart, portEnd, port) {
			return port
		}
	}
	LogErrorPanic(commandName, fmt.Sprintf("getFreePortFromRange() error - Cannot find free TCP ports in range %d-%d", portStart, portEnd))
	return 0
}

func IsPortFree(commandName string, portStart, portEnd int, port int) bool {
	tcpListen, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		slog.Debug(commandName, GetFuncName(), fmt.Sprintf("TCP %d port is reserved or already bound in range %d-%d", port, portStart, portEnd))
		return false
	}
	defer tcpListen.Close()

	return true
}

func IsHostnameExists(commandName string, hostname string) bool {
	_, err := net.LookupHost(hostname)
	if err != nil {
		slog.Debug(commandName, GetFuncName(), fmt.Sprintf("Host %s is unreacheable: %s", hostname, err.Error()))
	}

	return err == nil
}
