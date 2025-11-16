package keycloaksvc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/vaultclient"
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
	UpdateRealmAccessTokenSettings(lifespan int) error
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
	clientSecret := secrets[clientID].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenantName)
	systemUserPassword := secrets[systemUser].(string)

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

	return tokenData["access_token"].(string), nil
}

// TODO Add new tests
func (ks *KeycloakSvc) GetMasterAccessToken(grantType constant.KeycloakGrantType) (string, error) {
	formData := url.Values{}
	switch grantType {
	case constant.ClientCredentials:
		// TODO Extract literals to constant package
		formData.Set("grant_type", constant.ClientCredentials)
		formData.Set("client_id", "folio-backend-admin-client")
		formData.Set("client_secret", "supersecret")
	case constant.Password:
		// TODO Extract literals to constant package
		formData.Set("grant_type", constant.Password)
		formData.Set("client_id", "admin-cli")
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

	return tokenData["access_token"].(string), nil
}

// TODO Add tests
func (ks *KeycloakSvc) UpdateRealmAccessTokenSettings(lifespan int) error {
	payload, err := json.Marshal(map[string]any{
		"accessTokenLifespan": lifespan,
	})
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("%s/admin/realms/master", constant.KeycloakHTTP)
	headers := helpers.SecureApplicationJSONHeaders(ks.Action.KeycloakMasterAccessToken)
	if err := ks.HTTPClient.PutReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ks.Action.Name, "text", "Updated keycloak realm settings", "accessTokenLifespan", lifespan)

	return nil
}

func (ks *KeycloakSvc) UpdatePublicClientSettings(tenantName string, url string) error {
	clientID := fmt.Sprintf("%s%s", tenantName, action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", ks.Action.ConfigGlobalEnv))
	getRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", constant.KeycloakHTTP, tenantName, clientID)
	headers := helpers.SecureApplicationJSONHeaders(ks.Action.KeycloakMasterAccessToken)

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
