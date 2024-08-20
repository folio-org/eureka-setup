package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

// ####### HTTP ########

func DumpHttpBody(commandName string, enableDebug bool, bodyBytes []byte) {
	if !enableDebug {
		return
	}

	fmt.Println("### Dumping HTTP Request Body ###")
	fmt.Println(string(bodyBytes))
}

func DumpHttpRequest(commandName string, req *http.Request, enableDebug bool) {
	if !enableDebug {
		return
	}

	dumpResp, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error(commandName, MessageKey, "httputil.DumpRequest error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Request ###")
	fmt.Println(string(dumpResp))
}

func DumpHttpResponse(commandName string, resp *http.Response, enableDebug bool) {
	if !enableDebug {
		return
	}

	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error(commandName, MessageKey, "httputil.DumpResponse error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Response ###")
	fmt.Println(string(dumpResp))
}

// ####### STRING ########

func TrimModuleName(name string) string {
	if name[strings.LastIndex(name, "-")] == 45 {
		name = name[:strings.LastIndex(name, "-")]
	}

	return name
}

func TransformToEnvVar(name string) string {
	return EnvNameRegexp.ReplaceAllString(strings.ToUpper(name), `_`)
}
