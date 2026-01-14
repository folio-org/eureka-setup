package keycloaksvc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/httpclient"
	"github.com/j011195/eureka-setup/eureka-cli/managementsvc"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/vaultclient"
)

// KeycloakProcessor defines the interface for Keycloak service operations
type KeycloakProcessor interface {
	KeycloakAdminManager
	KeycloakUserManager
	KeycloakRoleManager
	KeycloakCapabilitySetManager
}

// KeycloakAdminManager defines the interface for Keycloak admin operations
type KeycloakAdminManager interface {
	GetAccessToken(tenantName string) (string, error)
	GetMasterAccessToken(grantType constant.KeycloakGrantType) (string, error)
	UpdateRealmAccessTokenSettings(tenantName string, lifespan int) error
	UpdatePublicClientSettings(tenantName string, url string) error
}

// KeycloakSvc provides functionality for Keycloak operations including user and role management
type KeycloakSvc struct {
	Action        *action.Action
	HTTPClient    httpclient.HTTPClientRunner
	VaultClient   vaultclient.VaultClientRunner
	ManagementSvc managementsvc.ManagementProcessor
}

// New creates a new KeycloakSvc instance
func New(action *action.Action,
	httpClient httpclient.HTTPClientRunner,
	vaultClient vaultclient.VaultClientRunner,
	managementSvc managementsvc.ManagementProcessor) *KeycloakSvc {
	return &KeycloakSvc{Action: action, HTTPClient: httpClient, VaultClient: vaultClient, ManagementSvc: managementSvc}
}

func (ks *KeycloakSvc) GetAccessToken(tenantName string) (string, error) {
	client, err := ks.VaultClient.Create()
	if err != nil {
		return "", err
	}

	secrets, err := ks.VaultClient.GetSecretKey(context.Background(), client, ks.Action.VaultRootToken, fmt.Sprintf("folio/%s", tenantName))
	if err != nil {
		return "", err
	}

	clientID := action.GetConfigEnv("KC_SERVICE_CLIENT_ID", ks.Action.ConfigGlobalEnv)
	clientSecret := helpers.GetString(secrets, clientID)
	systemUser := fmt.Sprintf("%s-system-user", tenantName)
	systemUserPassword := helpers.GetString(secrets, systemUser)

	formData := url.Values{}
	formData.Set("grant_type", constant.Password)
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	requestURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", constant.KeycloakHTTP, tenantName)
	headers := helpers.ApplicationFormURLEncodedHeaders()

	var tokenData map[string]any
	if err := ks.HTTPClient.PostFormDataReturnStruct(requestURL, formData, headers, &tokenData); err != nil {
		return "", err
	}
	if tokenData["access_token"] == nil {
		return "", errors.AccessTokenNotFound(requestURL)
	}

	return helpers.GetString(tokenData, "access_token"), nil
}

func (ks *KeycloakSvc) GetMasterAccessToken(grantType constant.KeycloakGrantType) (string, error) {
	formData := url.Values{}
	switch grantType {
	case constant.ClientCredentials:
		formData.Set("grant_type", constant.ClientCredentials)
		formData.Set("client_id", action.GetConfigEnv("KC_ADMIN_CLIENT_ID", ks.Action.ConfigGlobalEnv))
		formData.Set("client_secret", action.GetConfigEnv("KC_ADMIN_CLIENT_SECRET", ks.Action.ConfigGlobalEnv))
		formData.Set("scope", "email openid")
	case constant.Password:
		formData.Set("grant_type", constant.Password)
		formData.Set("client_id", constant.KeycloakAdminClient)
		formData.Set("username", constant.KeycloakAdminUsername)
		formData.Set("password", constant.KeycloakAdminPassword)
	}
	requestURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", constant.KeycloakHTTP)
	headers := helpers.ApplicationFormURLEncodedHeaders()

	var tokenData map[string]any
	if err := ks.HTTPClient.PostFormDataReturnStruct(requestURL, formData, headers, &tokenData); err != nil {
		return "", err
	}
	if tokenData["access_token"] == nil {
		return "", errors.AccessTokenNotFound(requestURL)
	}

	return helpers.GetString(tokenData, "access_token"), nil
}

func (ks *KeycloakSvc) UpdateRealmAccessTokenSettings(tenantName string, lifespan int) error {
	payload, err := json.Marshal(map[string]any{
		"accessTokenLifespan": lifespan,
	})
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("%s/admin/realms/%s", constant.KeycloakHTTP, tenantName)
	headers, err := helpers.SecureApplicationJSONHeaders(ks.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}
	if err := ks.HTTPClient.PutReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Updated keycloak realm settings", "realm", tenantName, "accessTokenLifespan", lifespan)

	return nil
}

func (ks *KeycloakSvc) UpdatePublicClientSettings(tenantName string, url string) error {
	clientID := fmt.Sprintf("%s%s", tenantName, action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", ks.Action.ConfigGlobalEnv))
	getRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", constant.KeycloakHTTP, tenantName, clientID)
	headers, err := helpers.SecureApplicationJSONHeaders(ks.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	var decodedResponse models.KeycloakClientsResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(getRequestURL, headers, &decodedResponse); err != nil {
		return err
	}
	if len(decodedResponse) != 1 {
		return errors.ClientNotFound(clientID)
	}

	clientUUID := decodedResponse[0].ID
	payload, err := json.Marshal(map[string]any{
		"rootUrl":                      url,
		"baseUrl":                      url,
		"adminUrl":                     url,
		"redirectUris":                 []string{fmt.Sprintf("%s/*", url)},
		"webOrigins":                   []string{"/*"},
		"authorizationServicesEnabled": true,
		"serviceAccountsEnabled":       true,
		"attributes": map[string]string{
			"post.logout.redirect.uris": fmt.Sprintf("%s/*", url),
			"login_theme":               "custom-theme",
		},
	})
	if err != nil {
		return err
	}

	putRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients/%s", constant.KeycloakHTTP, tenantName, clientUUID)
	if err := ks.HTTPClient.PutReturnNoContent(putRequestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Updated keycloak public client", "client", clientID, "realm", tenantName)

	return nil
}
