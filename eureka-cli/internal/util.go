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

	fmt.Printf("\n###### Dumping HTTP Request Body ######\n")
	fmt.Println(string(bodyBytes))
	fmt.Println()
}

func DumpHttpFormData(commandName string, enableDebug bool, formData url.Values) {
	if !enableDebug {
		return
	}

	fmt.Printf("\n###### Dumping HTTP Request Body ######\n")
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

	fmt.Printf("\n###### Dumping HTTP Request ######\n")
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

	fmt.Printf("\n###### Dumping HTTP Response ######\n")
	fmt.Println(string(respBytes))
	fmt.Println()
}

func CheckStatusCodes(commandName string, panicOnError bool, resp *http.Response) {
	if !panicOnError || resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}

	DumpHttpResponse(commandName, resp, true)

	LogErrorPanic(commandName, fmt.Sprintf("internal.CheckStatusCodes error - Unacceptable request status %d for URL: %s", resp.StatusCode, resp.Request.URL.String()))
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

func GetCurrentWorkDirPath(commandName string) string {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	return cwd
}

func GetHomeDirPath(commandName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.UserHomeDir error")
		panic(err)
	}

	homeDir := filepath.Join(home, ConfigDir)

	if err = os.MkdirAll(homeDir, 0644); err != nil {
		slog.Error(commandName, GetFuncName(), "os.MkdirAll error")
		panic(err)
	}

	return homeDir
}

// ######## Runtime ########

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)

	return runtime.FuncForPC(pc).Name()
}

// ######## Net ########

func GetAndSetFreePortFromRange(commandName string, portStart, portEnd int, excludeReservedPorts *[]int) int {
	for port := portStart; port <= portEnd; port++ {
		if !slices.Contains(*excludeReservedPorts, port) && IsPortFree(commandName, portStart, portEnd, port) {
			*excludeReservedPorts = append(*excludeReservedPorts, port)

			return port
		}
	}
	LogErrorPanic(commandName, fmt.Sprintf("internal.GetAndSetFreePortFromRange error - Cannot find free TCP ports in range %d-%d", portStart, portEnd))

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

func HostnameExists(commandName string, hostname string) bool {
	_, err := net.LookupHost(hostname)
	if err != nil {
		slog.Debug(commandName, GetFuncName(), fmt.Sprintf("Host %s is unreacheable: %s", hostname, err.Error()))
	}

	return err == nil
}
