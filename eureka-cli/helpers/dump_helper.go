package helpers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/folio-org/eureka-cli/action"
)

func DumpJSONBodyRequestBody(bodyBytes []byte) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return
	}

	fmt.Printf("\nDUMPING HTTP REQUEST BODY\n")
	fmt.Println(string(bodyBytes))
	fmt.Println()
}

func DumpFormDataRequestBody(formData url.Values) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return
	}

	fmt.Printf("\nDUMPING HTTP REQUEST BODY\n")
	fmt.Println(formData)
	fmt.Println()
}

func DumpRequest(action *action.Action, req *http.Request) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return
	}

	bytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	fmt.Printf("\nDUMPING HTTP REQUEST\n")
	fmt.Println(string(bytes))
	fmt.Println()
}

func DumpResponse(action *action.Action, resp *http.Response, forceDump bool) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) && !forceDump {
		return
	}

	bytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error(action.Name, "error", err)
		panic(err)
	}

	fmt.Printf("\nDUMPING HTTP RESPONSE\n")
	fmt.Println(string(bytes))
	fmt.Println()
}
