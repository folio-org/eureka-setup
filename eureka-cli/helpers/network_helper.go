package helpers

import (
	"fmt"
	"net"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
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

// ==================== Okapi Headers ====================

func SecureOkapiApplicationJSONHeaders(accessToken string) (map[string]string, error) {
	if accessToken == "" {
		return nil, errors.AccessTokenBlank()
	}

	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTokenHeader:  accessToken,
	}, nil
}

func SecureOkapiTenantApplicationJSONHeaders(tenantName string, accessToken string) (map[string]string, error) {
	if tenantName == "" {
		return nil, errors.TenantNameBlank()
	}
	if accessToken == "" {
		return nil, errors.AccessTokenBlank()
	}

	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenantName,
		constant.OkapiTokenHeader:  accessToken,
	}, nil
}

// ==================== Non-Okapi Headers ====================

func SecureTenantApplicationJSONHeaders(tenantName string, accessToken string) (map[string]string, error) {
	if tenantName == "" {
		return nil, errors.TenantNameBlank()
	}
	if accessToken == "" {
		return nil, errors.AccessTokenBlank()
	}

	return map[string]string{
		constant.ContentTypeHeader:   constant.ApplicationJSON,
		constant.OkapiTenantHeader:   tenantName,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}, nil
}

func SecureApplicationJSONHeaders(accessToken string) (map[string]string, error) {
	if accessToken == "" {
		return nil, errors.AccessTokenBlank()
	}

	return map[string]string{
		constant.ContentTypeHeader:   constant.ApplicationJSON,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}, nil
}

func ApplicationFormURLEncodedHeaders() map[string]string {
	return map[string]string{
		constant.ContentTypeHeader: constant.ApplicationFormURLEncoded,
	}
}

// ==================== Sidecar URL ====================

func GetSidecarURL(moduleName string, privatePort int) string {
	if strings.HasPrefix(moduleName, "edge") {
		return fmt.Sprintf("http://%s.eureka:%d", moduleName, privatePort)
	}

	return fmt.Sprintf("http://%s-sc.eureka:%d", moduleName, privatePort)
}
