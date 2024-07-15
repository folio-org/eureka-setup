package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
)

func DumpHttpRequest(commandName string, req *http.Request) {
	dumpResp, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "httputil.DumpRequest error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Request")
	fmt.Println(string(dumpResp))
}

func DumpHttpResponse(commandName string, resp *http.Response) {
	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "httputil.DumpResponse error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Response")
	fmt.Println(string(dumpResp))
}
