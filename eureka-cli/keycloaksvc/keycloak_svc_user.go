package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// KeycloakUserManager defines the interface for Keycloak user management operations
type KeycloakUserManager interface {
	GetUsers(tenantName string) ([]any, error)
	CreateUsers(configTenant string) error
	RemoveUsers(tenantName string) error
}

func (ks *KeycloakSvc) GetUsers(tenantName string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/users?offset=0&limit=10000")
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)

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
	for username, value := range ks.Action.ConfigUsers {
		entry := value.(map[string]any)
		tenantName := entry["tenant"].(string)
		if configTenant != tenantName {
			continue
		}

		createdUser, err := ks.createUser(tenantName, username, entry)
		if err != nil {
			return err
		}

		userID := createdUser["id"].(string)
		if err := ks.attachUserPassword(tenantName, userID, username, entry); err != nil {
			return err
		}

		userRoles := entry["roles"].([]any)
		if err := ks.attachUserRoles(tenantName, userID, username, userRoles); err != nil {
			return nil
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
			"firstName":              entry["first-name"].(string),
			"lastName":               entry["last-name"].(string),
			"email":                  fmt.Sprintf("%s-%s", tenantName, username),
			"preferredContactTypeId": "002",
		},
	})
	if err != nil {
		return nil, err
	}
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/users-keycloak/users")
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	// headers := helpers.SecureTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)

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
		"password": entry["password"].(string),
	})
	if err != nil {
		return err
	}

	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/authn/credentials")
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err := ks.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Attached password to user", "username", username, "tenant", tenantName)

	return nil
}

func (ks *KeycloakSvc) attachUserRoles(tenantName, userID, username string, userRoles []any) error {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles/users")
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)

	var roleIDs []string
	for _, userRole := range userRoles {
		role, err := ks.GetRoleByName(userRole.(string), headers)
		if err != nil {
			return err
		}

		roleID := role["id"].(string)
		if roleID == "" {
			slog.Warn(ks.Action.Name, "text", "Roles are not found", "role", role["name"].(string))
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

	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	for _, value := range users {
		entry := value.(map[string]any)
		username := entry["username"].(string)
		if ks.Action.ConfigUsers[username] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/users-keycloak/users/%s", entry["id"].(string)))
		if err := ks.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Removed user", "username", username, "tenant", tenantName)
	}

	return nil
}
