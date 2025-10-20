package keycloakstep

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (ks *KeycloakStep) GetRoles(headers map[string]string) ([]any, error) {
	requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, "/roles?offset=0&limit=10000")

	foundRolesMap, err := ks.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundRolesMap["roles"] == nil || len(foundRolesMap["roles"].([]any)) == 0 {
		return nil, err
	}

	return foundRolesMap["roles"].([]any), nil
}

func (ks *KeycloakStep) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/roles?query=name==%s", roleName))

	foundRolesMap, err := ks.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundRolesMap["roles"] == nil {
		return nil, nil
	}

	foundRoles := foundRolesMap["roles"].([]any)
	if len(foundRoles) != 1 {
		return nil, fmt.Errorf("number of found roles by %s role name is not 1", roleName)
	}

	return foundRoles[0].(map[string]any), nil
}

func (ks *KeycloakStep) CreateRoles(existingTenant string, accessToken string) error {
	requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, "/roles")
	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(field.Roles)

	for role, value := range roles {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		headers := map[string]string{
			constant.ContentTypeHeader: constant.ApplicationJSON,
			constant.OkapiTenantHeader: tenant,
			constant.OkapiTokenHeader:  accessToken,
		}

		b, err := json.Marshal(map[string]string{"name": caser.String(role), "description": "Default"})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.DoPostReturnNoContent(requestURL, b, headers)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Created %s role in %s tenant (realm)", role, tenant))
	}

	return nil
}

func (ks *KeycloakStep) RemoveRoles(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(field.Roles)

	foundRoles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	for _, value := range foundRoles {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if roles[roleName] == nil {
			continue
		}

		requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/roles/%s", mapEntry["id"].(string)))

		_ = ks.HTTPClient.DoDelete(requestURL, headers)

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Removed %s role in %s tenant (realm)", roleName, tenant))
	}

	return nil
}
