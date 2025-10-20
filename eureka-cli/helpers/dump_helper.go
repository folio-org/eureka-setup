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

func DumpRequestJSON(bodyBytes []byte) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return
	}

	fmt.Printf("\nDUMPING HTTP REQUEST BODY\n")
	fmt.Println(string(bodyBytes))
	fmt.Println()
}

func DumpRequestFormData(formData url.Values) {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return
	}

	fmt.Printf("\nDUMPING HTTP REQUEST BODY\n")
	fmt.Println(formData)
	fmt.Println()
}

func DumpRequest(action *action.Action, req *http.Request) error {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return nil
	}

	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}

	fmt.Printf("\nDUMPING HTTP REQUEST\n")
	fmt.Println(string(b))
	fmt.Println()

	return nil
}

func DumpResponse(action *action.Action, resp *http.Response, forceDump bool) error {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) && !forceDump {
		return nil
	}

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	fmt.Printf("\nDUMPING HTTP RESPONSE\n")
	fmt.Println(string(b))
	fmt.Println()

	return nil
}
