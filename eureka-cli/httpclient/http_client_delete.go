package httpclient

import (
	"bytes"
	"net/http"

	"github.com/folio-org/eureka-cli/helpers"
)

func (hc *HTTPClient) DoDelete(url string, headers map[string]string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := hc.ValidateResponse(resp); err != nil {
		return err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	return nil
}

func (hc *HTTPClient) DoDeleteWithBody(url string, b []byte, headers map[string]string) error {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := hc.ValidateResponse(resp); err != nil {
		return err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	return nil
}
