package helpers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func DumpRequest(httpRequest *http.Request) error {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return nil
	}

	payload, err := httputil.DumpRequest(httpRequest, true)
	if err != nil {
		return err
	}

	fmt.Printf("\nDUMPING HTTP REQUEST\n")
	fmt.Println(string(payload))
	fmt.Println()

	return nil
}

func DumpResponse(method, url string, httpResponse *http.Response, forceDump bool) error {
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) && !forceDump {
		return nil
	}

	payload, err := httputil.DumpResponse(httpResponse, true)
	if err != nil {
		return err
	}

	fmt.Printf("\nDUMPING HTTP RESPONSE\n")
	fmt.Printf("%s %s\n", method, url)
	fmt.Println(string(payload))
	fmt.Println()

	return nil
}
