package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
)

const (
	adminUsername string = "admin"
	adminPassword string = "admin"

	KeycloakUrl string = "http://keycloak.eureka:8080"
)

func GetKeycloakAccessToken(commandName string, enableDebug bool, vaultRootToken string, tenant string) string {
	secretMap := GetVaultSecretKey(commandName, enableDebug, vaultRootToken, fmt.Sprintf("folio/%s", tenant))

	clientId := GetEnvironmentFromMapByKey("KC_SERVICE_CLIENT_ID")
	clientSecret := secretMap[clientId].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenant)
	systemUserPassword := secretMap[systemUser].(string)
	requestUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", KeycloakUrl, tenant)
	headers := map[string]string{ContentTypeHeader: FormUrlEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", clientId)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	tokensMap := DoPostFormDataReturnMapStringAny(commandName, requestUrl, enableDebug, formData, headers)
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "internal.GetKeycloakAccessToken - Access token not found")
		return ""
	}

	return tokensMap["access_token"].(string)
}

func GetKeycloakMasterAccessToken(commandName string, enableDebug bool) string {
	requestUrl := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", KeycloakUrl)
	headers := map[string]string{ContentTypeHeader: FormUrlEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", "admin-cli")
	formData.Set("username", adminUsername)
	formData.Set("password", adminPassword)

	tokensMap := DoPostFormDataReturnMapStringAny(commandName, requestUrl, enableDebug, formData, headers)
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "internal.GetKeycloakAccessToken - Access token not found")
		return ""
	}

	return tokensMap["access_token"].(string)
}

func UpdateKeycloakPublicClientParams(commandName string, enableDebug bool, tenant string, accessToken string, platformCompleteUrl string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken)}

	clientId := fmt.Sprintf("%s%s", tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX"))
	getRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", KeycloakUrl, tenant, clientId)
	foundClients := DoGetDecodeReturnAny(commandName, getRequestUrl, enableDebug, true, headers).([]any)
	if len(foundClients) != 1 {
		LogErrorPanic(commandName, fmt.Sprintf("internal.UpdateKeycloakPublicClientParams - Number of found clients by %s client id is not 1", clientId))
		return
	}

	clientUuid := foundClients[0].(map[string]any)["id"].(string)
	clientParamsBytes, err := json.Marshal(map[string]any{
		"rootUrl":                      platformCompleteUrl,
		"baseUrl":                      platformCompleteUrl,
		"adminUrl":                     platformCompleteUrl,
		"redirectUris":                 []string{fmt.Sprintf("%s/*", platformCompleteUrl)},
		"webOrigins":                   []string{"/*"},
		"authorizationServicesEnabled": true,
		"serviceAccountsEnabled":       true,
		"attributes": map[string]string{
			"post.logout.redirect.uris": fmt.Sprintf("%s/*", platformCompleteUrl),
			"login_theme":               "custom-theme",
		},
	})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	putRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients/%s", KeycloakUrl, tenant, clientUuid)
	DoPutReturnNoContent(commandName, putRequestUrl, enableDebug, clientParamsBytes, headers)

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Updated keycloak public '%s' client in '%s' realm", clientId, tenant))
}
