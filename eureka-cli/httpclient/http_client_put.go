package httpclient

import (
	"net/http"
)

func (hc *HTTPClient) PutReturnNoContent(url string, b []byte, headers map[string]string) error {
	resp, err := hc.doRequest(http.MethodPut, url, b, headers, false)
	if err != nil {
		return err
	}
	defer CloseResponse(resp)

	return nil
}
