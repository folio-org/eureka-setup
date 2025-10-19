package keycloakstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/vaultclient"
)

type KeycloakStep struct {
	Action      *action.Action
	HTTPClient  *httpclient.HTTPClient
	VaultClient *vaultclient.VaultClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, vaultClient *vaultclient.VaultClient) *KeycloakStep {
	return &KeycloakStep{
		Action:      action,
		HTTPClient:  httpClient,
		VaultClient: vaultClient,
	}
}

func (ks *KeycloakStep) GetKeycloakAccessToken(vaultRootToken string, tenant string) string {
	client := ks.VaultClient.Create()

	secretMap := ks.VaultClient.GetSecretKey(client, vaultRootToken, fmt.Sprintf("folio/%s", tenant))

	clientID := helpers.GetConfigEnv("KC_SERVICE_CLIENT_ID")
	clientSecret := secretMap[clientID].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenant)
	systemUserPassword := secretMap[systemUser].(string)
	requestURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", constant.KeycloakHTTP, tenant)
	headers := map[string]string{constant.ContentTypeHeader: constant.FormURLEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	tokensMap := ks.HTTPClient.DoPostFormDataReturnMapStringAny(requestURL, formData, headers)
	if tokensMap["access_token"] == nil {
		helpers.LogErrorPanic(ks.Action, fmt.Errorf("access token not found from %s", requestURL))
		return ""
	}

	return tokensMap["access_token"].(string)
}

func (ks *KeycloakStep) GetKeycloakMasterAccessToken() string {
	requestURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", constant.KeycloakHTTP)

	headers := map[string]string{
		constant.ContentTypeHeader: constant.FormURLEncodedContentType,
	}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", "admin-cli")
	formData.Set("username", constant.KeycloakAdminUsername)
	formData.Set("password", constant.KeycloakAdminPassword)

	tokensMap := ks.HTTPClient.DoPostFormDataReturnMapStringAny(requestURL, formData, headers)
	if tokensMap["access_token"] == nil {
		helpers.LogErrorPanic(ks.Action, fmt.Errorf("access token not found from %s", requestURL))
		return ""
	}

	return tokensMap["access_token"].(string)
}

func (ks *KeycloakStep) UpdateKeycloakPublicClientParams(tenant string, accessToken string, url string) {
	headers := map[string]string{
		constant.ContentTypeHeader:   constant.JsonContentType,
		constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
	}

	clientID := fmt.Sprintf("%s%s", tenant, helpers.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX"))
	getRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", constant.KeycloakHTTP, tenant, clientID)
	foundClients := ks.HTTPClient.DoGetDecodeReturnAny(getRequestURL, true, headers).([]any)
	if len(foundClients) != 1 {
		helpers.LogErrorPanic(ks.Action, fmt.Errorf("number of found clients by %s client id is not 1", clientID))
		return
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
		slog.Error(ks.Action.Name, "error", err)
		panic(err)
	}

	putRequestURL := fmt.Sprintf("%s/admin/realms/%s/clients/%s", constant.KeycloakHTTP, tenant, clientUUID)
	ks.HTTPClient.DoPutReturnNoContent(putRequestURL, clientParamsBytes, headers)

	slog.Info(ks.Action.Name, "text", fmt.Sprintf("Updated keycloak public %s client in %s realm", clientID, tenant))
}
