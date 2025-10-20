package httpclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/hashicorp/go-retryablehttp"
)

func (hc *HTTPClient) DoPostReturnNoContent(url string, b []byte, headers map[string]string) error {
	helpers.DumpRequestJSON(b)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
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

func (hc *HTTPClient) DoRetryPostReturnNoContent(url string, b []byte, headers map[string]string) error {
	helpers.DumpRequestJSON(b)

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	SetRequestHeaders(req.Request, headers)
	helpers.DumpRequest(hc.Action, req.Request)

	resp, err := hc.retryClient.Do(req)
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

func (hc *HTTPClient) DoPostReturnMapStringAny(url string, b []byte, headers map[string]string) (map[string]any, error) {
	var respMap map[string]any

	helpers.DumpRequestJSON(b)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := hc.ValidateResponse(resp); err != nil {
		return nil, err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

func (hc *HTTPClient) DoPostFormDataReturnMapStringAny(url string, fd url.Values, headers map[string]string) (map[string]any, error) {
	var respMap map[string]any

	helpers.DumpRequestFormData(fd)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(fd.Encode()))
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := hc.ValidateResponse(resp); err != nil {
		return nil, err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	err = json.NewDecoder(resp.Body).Decode(&respMap)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}
