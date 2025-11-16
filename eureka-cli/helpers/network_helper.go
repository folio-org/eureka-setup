package helpers

import (
	"fmt"
	"net"
	"strings"

	"github.com/folio-org/eureka-cli/constant"
)

// ==================== Hostname ====================

func IsHostnameReachable(actionName string, hostname string) error {
	_, err := net.LookupHost(hostname)
	if err != nil {
		return err
	}

	return nil
}

// ==================== Hostname ====================

func ConstructURL(url string, gatewayURL string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	return fmt.Sprintf("%s:%s", gatewayURL, url)
}

func ExtractPortFromURL(url string) (int, error) {
	port, err := GetPortFromURL(url)
	if err != nil {
		return 0, err
	}

	return port, nil
}

// ==================== Okapi Headers ====================

// TODO Check if accessToken is not blank
func SecureOkapiApplicationJSONHeaders(accessToken string) map[string]string {
	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTokenHeader:  accessToken,
	}
}

// TODO Check if tenantName or accessToken are not blank
func SecureOkapiTenantApplicationJSONHeaders(tenantName string, accessToken string) map[string]string {
	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenantName,
		constant.OkapiTokenHeader:  accessToken,
	}
}

// ==================== Non-Okapi Headers ====================

// TODO Check if tenantName or accessToken are not blank
func SecureTenantApplicationJSONHeaders(tenantName string, accessToken string) map[string]string {
	return map[string]string{
		constant.ContentTypeHeader:   constant.ApplicationJSON,
		constant.OkapiTenantHeader:   tenantName,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}
}

// TODO Check if accessToken is not blank
func SecureApplicationJSONHeaders(accessToken string) map[string]string {
	return map[string]string{
		constant.ContentTypeHeader:   constant.ApplicationJSON,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}
}

func ApplicationFormURLEncodedHeaders() map[string]string {
	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationFormURLEncoded,
	}
}
