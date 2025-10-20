package httpclient

import (
	"bytes"
	"net/http"

	"github.com/folio-org/eureka-cli/helpers"
)

func (hc *HTTPClient) DoPutReturnNoContent(url string, b []byte, headers map[string]string) error {
	helpers.DumpRequestJSON(b)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
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
