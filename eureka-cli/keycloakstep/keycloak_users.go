package keycloakstep

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func (ks *KeycloakStep) GetUsers(tenant string, accessToken string) ([]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/users?offset=0&limit=10000")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundUsersMap, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundUsersMap["users"] == nil || len(foundUsersMap["users"].([]any)) == 0 {
		return nil, nil
	}

	return foundUsersMap["users"].([]any), nil
}

func (ks *KeycloakStep) CreateUsers(panicOnError bool, existingTenant string, accessToken string) error {
	postUserRequestURL := ks.Action.CreateURL(constant.KongPort, "/users-keycloak/users")
	postUserPasswordRequestURL := ks.Action.CreateURL(constant.KongPort, "/authn/credentials")
	postUserRoleRequestURL := ks.Action.CreateURL(constant.KongPort, "/roles/users")
	usersMap := viper.GetStringMap(field.Users)

	for username, value := range usersMap {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		password := mapEntry["password"].(string)
		firstName := mapEntry["first-name"].(string)
		lastName := mapEntry["last-name"].(string)
		userRoles := mapEntry["roles"].([]any)

		userBytes, err := json.Marshal(map[string]any{
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

		createdUserMap, err := ks.HTTPClient.PostReturnMapStringAny(postUserRequestURL, userBytes, okapiBasedHeaders)
		if err != nil {
			return err
		}

		userId := createdUserMap["id"].(string)

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Created %s user with password %s in %s tenant (realm)", username, password, tenant))

		userPasswordBytes, err := json.Marshal(map[string]any{"userId": userId, "username": username, "password": password})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(postUserPasswordRequestURL, userPasswordBytes, nonOkapiBasedHeaders)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Attached %s password to %s user in %s tenant (realm)", password, username, tenant))

		var roleIds []string
		for _, userRole := range userRoles {
			role, err := ks.GetRoleByName(userRole.(string), okapiBasedHeaders)
			if err != nil {
				return err
			}

			roleId := role["id"].(string)
			roleName := role["name"].(string)

			if roleId == "" {
				slog.Warn(ks.Action.Name, "text", fmt.Sprintf("did not find role %s by name", roleName))
				continue
			}

			roleIds = append(roleIds, roleId)
		}

		userRoleBytes, err := json.Marshal(map[string]any{"userId": userId, "roleIds": roleIds})
		if err != nil {
			return err
		}

		ks.HTTPClient.PostReturnNoContent(postUserRoleRequestURL, userRoleBytes, okapiBasedHeaders)

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Attached %d roles to %s user in %s tenant (realm)", len(roleIds), username, tenant))
	}

	return nil
}

func (ks *KeycloakStep) RemoveUsers(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundUsers, err := ks.GetUsers(tenant, accessToken)
	if err != nil {
		return err
	}

	for _, value := range foundUsers {
		mapEntry := value.(map[string]any)

		username := mapEntry["username"].(string)
		usersMap := viper.GetStringMap(field.Users)
		if usersMap[username] == nil {
			continue
		}

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		_ = ks.HTTPClient.Delete(requestURL, headers)

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Removed %s in %s tenant (realm)", username, tenant))
	}

	return nil
}
