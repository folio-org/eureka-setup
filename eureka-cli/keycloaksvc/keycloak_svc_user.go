package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
)

// KeycloakUserManager defines the interface for Keycloak user management operations
type KeycloakUserManager interface {
	GetUsers(tenantName string) ([]any, error)
	CreateUsers(configTenant string) error
	RemoveUsers(tenantName string) error
}

func (ks *KeycloakSvc) GetUsers(tenantName string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/users?offset=0&limit=10000")
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return nil, err
	}

	var decodedResponse models.KeycloakUsersResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}
	if len(decodedResponse.Users) == 0 {
		return nil, nil
	}

	result := make([]any, len(decodedResponse.Users))
	for i, user := range decodedResponse.Users {
		result[i] = map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"active":   user.Active,
			"type":     user.Type,
			"personal": user.Personal,
		}
	}

	return result, nil
}

func (ks *KeycloakSvc) CreateUsers(configTenant string) error {
	usernames := make([]string, 0, len(ks.Action.ConfigUsers))
	for username := range ks.Action.ConfigUsers {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	for _, username := range usernames {
		value := ks.Action.ConfigUsers[username]
		entry := value.(map[string]any)
		tenantName := helpers.GetString(entry, "tenant")
		if configTenant != tenantName {
			continue
		}

		createdUser, err := ks.createUser(tenantName, username, entry)
		if err != nil {
			return err
		}

		userID := helpers.GetString(createdUser, "id")
		if err := ks.attachUserPassword(tenantName, userID, username, entry); err != nil {
			return err
		}

		userRoles := helpers.GetAnySlice(entry, "roles")
		if len(userRoles) > 0 {
			if err := ks.attachUserRoles(tenantName, userID, username, userRoles); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ks *KeycloakSvc) createUser(tenantName string, username string, entry map[string]any) (map[string]any, error) {
	payload, err := json.Marshal(map[string]any{
		"username": username,
		"active":   true,
		"type":     "staff",
		"personal": map[string]any{
			"firstName":              helpers.GetString(entry, "first-name"),
			"lastName":               helpers.GetString(entry, "last-name"),
			"email":                  fmt.Sprintf("%s_%s@test.org", tenantName, username),
			"preferredContactTypeId": "002",
		},
	})
	if err != nil {
		return nil, err
	}
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/users-keycloak/users")
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return nil, err
	}

	var decodedResponse map[string]any
	if err := ks.HTTPClient.PostReturnStruct(requestURL, payload, headers, &decodedResponse); err != nil {
		return nil, err
	}
	slog.Info(ks.Action.Name, "text", "Created user with password", "username", username, "tenant", tenantName)

	return decodedResponse, nil
}

func (ks *KeycloakSvc) attachUserPassword(tenantName, userID, username string, entry map[string]any) error {
	payload, err := json.Marshal(map[string]any{
		"userId":   userID,
		"username": username,
		"password": helpers.GetString(entry, "password"),
	})
	if err != nil {
		return err
	}

	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/authn/credentials")
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}
	if err := ks.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Attached password to user", "username", username, "tenant", tenantName)

	return nil
}

func (ks *KeycloakSvc) attachUserRoles(tenantName, userID, username string, userRoles []any) error {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles/users")
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

	var roleIDs []string
	for _, userRole := range userRoles {
		role, err := ks.GetRoleByName(userRole.(string), headers)
		if err != nil {
			return err
		}

		roleID := helpers.GetString(role, "id")
		if roleID == "" {
			slog.Warn(ks.Action.Name, "text", "Roles are not found", "username", username, "role", helpers.GetString(role, "name"))
			continue
		}
		roleIDs = append(roleIDs, roleID)
	}

	payload, err := json.Marshal(map[string]any{
		"userId":  userID,
		"roleIds": roleIDs,
	})
	if err != nil {
		return err
	}
	if err := ks.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Attached roles to user", "username", username, "tenant", tenantName, "count", len(roleIDs))

	return nil
}

func (ks *KeycloakSvc) RemoveUsers(tenantName string) error {
	users, err := ks.GetUsers(tenantName)
	if err != nil {
		return err
	}

	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

	for _, value := range users {
		entry := value.(map[string]any)
		username := helpers.GetString(entry, "username")
		if ks.Action.ConfigUsers[username] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/users-keycloak/users/%s", helpers.GetString(entry, "id")))
		if err := ks.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed user", "username", username, "tenant", tenantName)
	}

	return nil
}
