package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/helpers"
)

// KeycloakRoleManager defines the interface for Keycloak role management operations
type KeycloakRoleManager interface {
	GetRoles(headers map[string]string) ([]any, error)
	GetRoleByName(roleName string, headers map[string]string) (map[string]any, error)
	CreateRoles(configTenant string) error
	RemoveRoles(tenantName string) error
}

func (ks *KeycloakSvc) GetRoles(headers map[string]string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles?offset=0&limit=10000")
	decodedResponse, err := ks.HTTPClient.GetRetryDecodeReturnAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	rolesData := decodedResponse.(map[string]any)
	if rolesData["roles"] == nil || len(rolesData["roles"].([]any)) == 0 {
		return nil, err
	}

	return rolesData["roles"].([]any), nil
}

func (ks *KeycloakSvc) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles?query=name==%s", roleName))
	decodedResponse, err := ks.HTTPClient.GetRetryDecodeReturnAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	rolesData := decodedResponse.(map[string]any)
	if rolesData["roles"] == nil {
		return nil, nil
	}

	roles := rolesData["roles"].([]any)
	if len(roles) != 1 {
		return nil, errors.RoleNotFound(roleName)
	}

	return roles[0].(map[string]any), nil
}

func (ks *KeycloakSvc) CreateRoles(configTenant string) error {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles")
	for role, value := range ks.Action.ConfigRoles {
		entry := value.(map[string]any)
		tenantName := entry["tenant"].(string)
		if configTenant != tenantName {
			continue
		}

		headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
		payload, err := json.Marshal(map[string]string{
			"name":        ks.Action.Caser.String(role),
			"description": "Default",
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(requestURL, payload, headers)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Created role in tenant", "role", role, "tenant", tenantName)
	}

	return nil
}

func (ks *KeycloakSvc) RemoveRoles(tenantName string) error {
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	for _, value := range roles {
		entry := value.(map[string]any)
		roleName := ks.Action.Caser.String(entry["name"].(string))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s", entry["id"].(string)))
		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed role in tenant", "role", roleName, "tenant", tenantName)
	}

	return nil
}
