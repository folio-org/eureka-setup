package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/spf13/viper"
)

const (
	adminUsername string = "admin"
	adminPassword string = "admin"
)

func GetKeycloakAccessToken(commandName string, enableDebug bool, vaultRootToken string, tenant string) string {
	secretMap := GetVaultSecretKey(commandName, enableDebug, vaultRootToken, fmt.Sprintf("folio/%s", tenant))

	clientId := GetEnvironmentFromMapByKey("KC_SERVICE_CLIENT_ID")
	clientSecret := secretMap[clientId].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenant)
	systemUserPassword := secretMap[systemUser].(string)
	requestUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", viper.GetString(ResourcesKeycloakKey), tenant)
	headers := map[string]string{ContentTypeHeader: FormUrlEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", clientId)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	tokensMap := DoPostFormDataReturnMapStringInteface(commandName, requestUrl, enableDebug, formData, headers)
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "internal.GetKeycloakAccessToken - Access token not found")
	}

	return tokensMap["access_token"].(string)
}

func GetKeycloakMasterRealmAccessToken(commandName string, enableDebug bool) string {
	requestUrl := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", viper.GetString(ResourcesKeycloakKey))
	headers := map[string]string{ContentTypeHeader: FormUrlEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", "admin-cli")
	formData.Set("username", adminUsername)
	formData.Set("password", adminPassword)

	tokensMap := DoPostFormDataReturnMapStringInteface(commandName, requestUrl, enableDebug, formData, headers)
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "internal.GetKeycloakAccessToken - Access token not found")
	}

	return tokensMap["access_token"].(string)
}

func UpdateKeycloakPublicClientParams(commandName string, enableDebug bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken)}

	clientId := fmt.Sprintf("%s%s", tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX"))
	getRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", viper.GetString(ResourcesKeycloakKey), tenant, clientId)
	foundClients := DoGetDecodeReturnInterface(commandName, getRequestUrl, enableDebug, true, headers).([]interface{})
	if len(foundClients) != 1 {
		LogErrorPanic(commandName, fmt.Sprintf("internal.UpdateKeycloakPublicClientParams - Number of found cliends by %s client id is not 1", clientId))
	}

	clientUuid := foundClients[0].(map[string]interface{})["id"].(string)

	putRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients/%s", viper.GetString(ResourcesKeycloakKey), tenant, clientUuid)
	clientParamsBytes, err := json.Marshal(map[string]interface{}{
		"rootUrl":                      PlatformCompleteUrl,
		"baseUrl":                      PlatformCompleteUrl,
		"adminUrl":                     PlatformCompleteUrl,
		"redirectUris":                 []string{fmt.Sprintf("%s/*", PlatformCompleteUrl)},
		"webOrigins":                   []string{"/*"},
		"authorizationServicesEnabled": true,
		"serviceAccountsEnabled":       true,
		"attributes": map[string]string{
			"post.logout.redirect.uris": fmt.Sprintf("%s/*", PlatformCompleteUrl),
			"login_theme":               "custom-theme",
		},
	})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	DoPutReturnNoContent(commandName, putRequestUrl, enableDebug, clientParamsBytes, headers)
}
