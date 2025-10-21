package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	managementsvc "github.com/folio-org/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-cli/vaultclient"
)

type KeycloakSvc struct {
	Action        *action.Action
	HTTPClient    *httpclient.HTTPClient
	VaultClient   *vaultclient.VaultClient
	ManagementSvc *managementsvc.ManagementSvc
}

func New(
	action *action.Action,
	httpClient *httpclient.HTTPClient,
	vaultClient *vaultclient.VaultClient,
	managementSvc *managementsvc.ManagementSvc,
) *KeycloakSvc {
	return &KeycloakSvc{
		Action:        action,
		HTTPClient:    httpClient,
		VaultClient:   vaultClient,
		ManagementSvc: managementSvc,
	}
}

func (ks *KeycloakSvc) GetKeycloakAccessToken(vaultRootToken string, tenant string) (string, error) {
	client, err := ks.VaultClient.Create()
	if err != nil {
		return "", err
	}

	secretMap, err := ks.VaultClient.GetSecretKey(client, vaultRootToken, fmt.Sprintf("folio/%s", tenant))
	if err != nil {
		return "", err
	}

	clientID := helpers.GetConfigEnv("KC_SERVICE_CLIENT_ID")
	clientSecret := secretMap[clientID].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenant)
	systemUserPassword := secretMap[systemUser].(string)
	requestURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", constant.KeycloakHTTP, tenant)
	headers := map[string]string{constant.ContentTypeHeader: constant.ApplicationFormURLEncoded}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	tokensMap, err := ks.HTTPClient.PostFormDataReturnMapStringAny(requestURL, formData, headers)
	if err != nil {
		return "", err
	}

	if tokensMap["access_token"] == nil {
		return "", fmt.Errorf("access token not found from %s", requestURL)
	}

	return tokensMap["access_token"].(string), nil
}

func (ks *KeycloakSvc) GetKeycloakMasterAccessToken() (string, error) {
	requestURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", constant.KeycloakHTTP)

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationFormURLEncoded,
	}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", "admin-cli")
	formData.Set("username", constant.KeycloakAdminUsername)
	formData.Set("password", constant.KeycloakAdminPassword)

	tokensMap, err := ks.HTTPClient.PostFormDataReturnMapStringAny(requestURL, formData, headers)
	if err != nil {
		return "", err
	}

	if tokensMap["access_token"] == nil {
		return "", fmt.Errorf("access token not found from %s", requestURL)
	}

	return tokensMap["access_token"].(string), nil
}

func (ks *KeycloakSvc) UpdateKeycloakPublicClientParams(tenant string, accessToken string, url string) error {
	headers := map[string]string{
		constant.ContentTypeHeader:   constant.ApplicationJSON,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}

	clientID := fmt.Sprintf("%s%s", tenant, helpers.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX"))
	getRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", constant.KeycloakHTTP, tenant, clientID)
	foundClientsResp, err := ks.HTTPClient.GetRetryDecodeReturnAny(getRequestURL, headers)
	if err != nil {
		return err
	}

	foundClients := foundClientsResp.([]any)

	if len(foundClients) != 1 {
		return fmt.Errorf("number of found clients by %s client id is not 1", clientID)
	}

	clientUUID := foundClients[0].(map[string]any)["id"].(string)
	clientParamsBytes, err := json.Marshal(map[string]any{
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

	putRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients/%s", constant.KeycloakHTTP, tenant, clientUUID)
	err = ks.HTTPClient.PutReturnNoContent(putRequestURL, clientParamsBytes, headers)
	if err != nil {
		return err
	}

	slog.Info(ks.Action.Name, "text", "Updated keycloak public client in realm", "client", clientID, "realm", tenant)

	return nil
}
