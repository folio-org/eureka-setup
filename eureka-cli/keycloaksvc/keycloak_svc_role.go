package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
)

type KeycloakRoleManager interface {
	GetRoles(headers map[string]string) ([]any, error)
	GetRoleByName(roleName string, headers map[string]string) (map[string]any, error)
	CreateRoles(configTenant string) error
	RemoveRoles(tenantName string) error
}

func (ks *KeycloakSvc) GetRoles(headers map[string]string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles?offset=0&limit=10000")
	resp, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}
	if resp["roles"] == nil || len(resp["roles"].([]any)) == 0 {
		return nil, err
	}

	return resp["roles"].([]any), nil
}

func (ks *KeycloakSvc) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles?query=name==%s", roleName))
	resp, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}
	if resp["roles"] == nil {
		return nil, nil
	}

	roles := resp["roles"].([]any)
	if len(roles) != 1 {
		return nil, fmt.Errorf("number of found roles by %s role name is not 1", roleName)
	}

	return roles[0].(map[string]any), nil
}

func (ks *KeycloakSvc) CreateRoles(configTenant string) error {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles")
	for role, value := range ks.Action.ConfigRoles {
		mapEntry := value.(map[string]any)
		tenantName := mapEntry["tenant"].(string)
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
		mapEntry := value.(map[string]any)
		roleName := ks.Action.Caser.String(mapEntry["name"].(string))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s", mapEntry["id"].(string)))
		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed role in tenant", "role", roleName, "tenant", tenantName)
	}

	return nil
}
