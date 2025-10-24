package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func (ks *KeycloakSvc) GetUsers(tenant string, accessToken string) ([]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/users?offset=0&limit=10000")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	uu, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if uu["users"] == nil || len(uu["users"].([]any)) == 0 {
		return nil, nil
	}

	return uu["users"].([]any), nil
}

func (ks *KeycloakSvc) CreateUsers(existingTenant string, accessToken string) error {
	postUserRequestURL := ks.Action.CreateURL(constant.KongPort, "/users-keycloak/users")
	postUserPasswordRequestURL := ks.Action.CreateURL(constant.KongPort, "/authn/credentials")
	postUserRoleRequestURL := ks.Action.CreateURL(constant.KongPort, "/roles/users")
	uu1 := viper.GetStringMap(field.Users)

	for username, value := range uu1 {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		password := mapEntry["password"].(string)
		firstName := mapEntry["first-name"].(string)
		lastName := mapEntry["last-name"].(string)
		userRoles := mapEntry["roles"].([]any)

		bb1, err := json.Marshal(map[string]any{
			"username": username,
			"active":   true,
			"type":     "staff",
			"personal": map[string]any{
				"firstName":              firstName,
				"lastName":               lastName,
				"email":                  fmt.Sprintf("%s-%s", tenant, username),
				"preferredContactTypeId": "002",
			},
		})
		if err != nil {
			return err
		}

		okapiBasedHeaders := map[string]string{
			constant.ContentTypeHeader: constant.ApplicationJSON,
			constant.OkapiTenantHeader: tenant,
			constant.OkapiTokenHeader:  accessToken,
		}

		nonOkapiBasedHeaders := map[string]string{
			constant.ContentTypeHeader:   constant.ApplicationJSON,
			constant.OkapiTenantHeader:   tenant,
			constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
		}

		uu2, err := ks.HTTPClient.PostReturnMapStringAny(postUserRequestURL, bb1, okapiBasedHeaders)
		if err != nil {
			return err
		}

		userID := uu2["id"].(string)

		slog.Info(ks.Action.Name, "text", "Created user with password in tenant", "username", username, "password", password, "tenant", tenant)

		bb2, err := json.Marshal(map[string]any{
			"userId":   userID,
			"username": username,
			"password": password,
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(postUserPasswordRequestURL, bb2, nonOkapiBasedHeaders)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Attached password to user in tenant", "password", password, "username", username, "tenant", tenant)

		var roleIDs []string
		for _, userRole := range userRoles {
			role, err := ks.GetRoleByName(userRole.(string), okapiBasedHeaders)
			if err != nil {
				return err
			}

			roleID := role["id"].(string)
			roleName := role["name"].(string)

			if roleID == "" {
				slog.Warn(ks.Action.Name, "text", "Did not find role by name", "role", roleName)
				continue
			}

			roleIDs = append(roleIDs, roleID)
		}

		bb3, err := json.Marshal(map[string]any{
			"userId":  userID,
			"roleIds": roleIDs,
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(postUserRoleRequestURL, bb3, okapiBasedHeaders)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Attached roles to user in tenant", "roleCount", len(roleIDs), "username", username, "tenant", tenant)
	}

	return nil
}

func (ks *KeycloakSvc) RemoveUsers(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	uu, err := ks.GetUsers(tenant, accessToken)
	if err != nil {
		return err
	}

	for _, value := range uu {
		mapEntry := value.(map[string]any)

		username := mapEntry["username"].(string)
		usersMap := viper.GetStringMap(field.Users)
		if usersMap[username] == nil {
			continue
		}

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Removed user in tenant", "username", username, "tenant", tenant)
	}

	return nil
}
