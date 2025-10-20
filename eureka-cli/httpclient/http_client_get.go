package httpclient

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/folio-org/eureka-cli/helpers"
)

func (hc *HTTPClient) DoGetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return nil, err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	return resp, nil
}

func (hc *HTTPClient) DoGetDecodeReturnString(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	SetRequestHeaders(req, headers)
	helpers.DumpRequest(hc.Action, req)

	resp, err := hc.customClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := hc.ValidateResponse(resp); err != nil {
		return "", err
	}

	helpers.DumpResponse(hc.Action, resp, false)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (hc *HTTPClient) DoGetDecodeReturnAny(url string, headers map[string]string) (any, error) {
	var respMap any

	req, err := http.NewRequest(http.MethodGet, url, nil)
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

func (hc *HTTPClient) DoGetDecodeReturnMapStringAny(url string, headers map[string]string) (map[string]any, error) {
	var respMap map[string]any

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
