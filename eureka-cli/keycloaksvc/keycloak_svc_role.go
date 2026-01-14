package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
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

	var decodedResponse models.KeycloakRolesResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}

	if len(decodedResponse.Roles) == 0 {
		return nil, nil
	}

	result := make([]any, len(decodedResponse.Roles))
	for i, role := range decodedResponse.Roles {
		result[i] = map[string]any{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
		}
	}

	return result, nil
}

func (ks *KeycloakSvc) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles?query=name==%s&limit=1", roleName))

	var decodedResponse models.KeycloakRolesResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}
	if len(decodedResponse.Roles) == 0 {
		return nil, nil
	}
	if len(decodedResponse.Roles) != 1 {
		return nil, errors.RoleNotFound(roleName)
	}

	return map[string]any{
		"id":          decodedResponse.Roles[0].ID,
		"name":        decodedResponse.Roles[0].Name,
		"description": decodedResponse.Roles[0].Description,
	}, nil
}

func (ks *KeycloakSvc) CreateRoles(configTenant string) error {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles")
	for role, value := range ks.Action.ConfigRoles {
		entry := value.(map[string]any)
		tenantName := helpers.GetString(entry, "tenant")
		if configTenant != tenantName {
			continue
		}

		headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]string{
			"name":        ks.Action.Caser.String(role),
			"description": "Default",
		})
		if err != nil {
			return err
		}
		if err := ks.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Created role", "role", role, "tenant", tenantName)
	}

	return nil
}

func (ks *KeycloakSvc) RemoveRoles(tenantName string) error {
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	for _, value := range roles {
		entry := value.(map[string]any)
		roleName := ks.Action.Caser.String(helpers.GetString(entry, "name"))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s", helpers.GetString(entry, "id")))
		if err := ks.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed role", "role", roleName, "tenant", tenantName)
	}

	return nil
}
