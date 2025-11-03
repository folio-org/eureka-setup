package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
)

// KeycloakUserManager defines the interface for Keycloak user management operations
type KeycloakUserManager interface {
	GetUsers(tenantName string) ([]any, error)
	CreateUsers(configTenant string) error
	RemoveUsers(tenantName string) error
}

func (ks *KeycloakSvc) GetUsers(tenantName string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/users?offset=0&limit=10000")
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	decodedResponse, err := ks.HTTPClient.GetRetryDecodeReturnAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	usersData := decodedResponse.(map[string]any)
	if usersData["users"] == nil || len(usersData["users"].([]any)) == 0 {
		return nil, nil
	}

	return usersData["users"].([]any), nil
}

func (ks *KeycloakSvc) CreateUsers(configTenant string) error {
	postUserRequestURL := ks.Action.GetRequestURL(constant.KongPort, "/users-keycloak/users")
	postUserPasswordRequestURL := ks.Action.GetRequestURL(constant.KongPort, "/authn/credentials")
	postUserRoleRequestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles/users")
	for username, value := range ks.Action.ConfigUsers {
		mapEntry := value.(map[string]any)
		tenantName := mapEntry["tenant"].(string)
		if configTenant != tenantName {
			continue
		}

		password := mapEntry["password"].(string)
		firstName := mapEntry["first-name"].(string)
		lastName := mapEntry["last-name"].(string)
		userRoles := mapEntry["roles"].([]any)
		payload1, err := json.Marshal(map[string]any{
			"username": username,
			"active":   true,
			"type":     "staff",
			"personal": map[string]any{
				"firstName":              firstName,
				"lastName":               lastName,
				"email":                  fmt.Sprintf("%s-%s", tenantName, username),
				"preferredContactTypeId": "002",
			},
		})
		if err != nil {
			return err
		}

		okapiBasedHeaders := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
		nonOkapiBasedHeaders := helpers.NonOkapiSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
		var createdUser map[string]any
		err = ks.HTTPClient.PostReturnStruct(postUserRequestURL, payload1, okapiBasedHeaders, &createdUser)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Created user with password in tenant", "username", username, "password", password, "tenant", tenantName)
		userID := createdUser["id"].(string)
		payload2, err := json.Marshal(map[string]any{
			"userId":   userID,
			"username": username,
			"password": password,
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(postUserPasswordRequestURL, payload2, nonOkapiBasedHeaders)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Attached password to user in tenant", "password", password, "username", username, "tenant", tenantName)

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

		payload3, err := json.Marshal(map[string]any{
			"userId":  userID,
			"roleIds": roleIDs,
		})
		if err != nil {
			return err
		}

		err = ks.HTTPClient.PostReturnNoContent(postUserRoleRequestURL, payload3, okapiBasedHeaders)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Attached roles to user in tenant", "roleCount", len(roleIDs), "username", username, "tenant", tenantName)
	}

	return nil
}

func (ks *KeycloakSvc) RemoveUsers(tenantName string) error {
	users, err := ks.GetUsers(tenantName)
	if err != nil {
		return err
	}

	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	for _, value := range users {
		mapEntry := value.(map[string]any)
		username := mapEntry["username"].(string)
		if ks.Action.ConfigUsers[username] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))
		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed user in tenant", "username", username, "tenant", tenantName)
	}

	return nil
}
