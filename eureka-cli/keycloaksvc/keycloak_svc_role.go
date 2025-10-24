package keycloaksvc

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

func (ks *KeycloakSvc) GetRoles(headers map[string]string) ([]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/roles?offset=0&limit=10000")

	rr, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if rr["roles"] == nil || len(rr["roles"].([]any)) == 0 {
		return nil, err
	}

	return rr["roles"].([]any), nil
}

func (ks *KeycloakSvc) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/roles?query=name==%s", roleName))

	rr1, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if rr1["roles"] == nil {
		return nil, nil
	}

	rr2 := rr1["roles"].([]any)
	if len(rr2) != 1 {
		return nil, fmt.Errorf("number of found roles by %s role name is not 1", roleName)
	}

	return rr2[0].(map[string]any), nil
}

func (ks *KeycloakSvc) CreateRoles(existingTenant string, accessToken string) error {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/roles")
	caser := cases.Lower(language.English)
	rr := viper.GetStringMap(field.Roles)

	for role, value := range rr {
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

		bb, err := json.Marshal(map[string]string{
			"name":        caser.String(role),
			"description": "Default",
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(requestURL, bb, headers)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Created role in tenant", "role", role, "tenant", tenant)
	}

	return nil
}

func (ks *KeycloakSvc) RemoveRoles(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(field.Roles)

	rr, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	for _, value := range rr {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if roles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/roles/%s", mapEntry["id"].(string)))

		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Removed role in tenant", "role", roleName, "tenant", tenant)
	}

	return nil
}
