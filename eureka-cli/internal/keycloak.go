package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/spf13/viper"
)

const (
	platformCompleteUrl string = "http://localhost:3000"
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
		LogErrorPanic(commandName, "")
	}

	return tokensMap["access_token"].(string)
}

func GetKeycloakMasterRealmAccessToken(commandName string, enableDebug bool) string {
	adminCliRequestUrl := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", viper.GetString(ResourcesKeycloakKey))
	adminCliHeaders := map[string]string{ContentTypeHeader: FormUrlEncodedContentType}

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", "admin-cli")
	formData.Set("username", "admin")
	formData.Set("password", "admin")

	tokensMap := DoPostFormDataReturnMapStringInteface(commandName, adminCliRequestUrl, enableDebug, formData, adminCliHeaders)
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "")
	}

	return tokensMap["access_token"].(string)
}

func UpdateKeycloakPublicClientParams(commandName string, enableDebug bool, tenant string, accessToken string) {
	clientId := fmt.Sprintf("%s%s", tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX"))
	clientGetRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", viper.GetString(ResourcesKeycloakKey), tenant, clientId)
	clientHeaders := map[string]string{ContentTypeHeader: JsonContentType, AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken)}
	clientUuid := DoGetDecodeReturnInterface(commandName, clientGetRequestUrl, enableDebug, true, clientHeaders).([]interface{})[0].(map[string]interface{})["id"].(string)

	clientPutRequestUrl := fmt.Sprintf("%s/admin/realms/%s/clients/%s", viper.GetString(ResourcesKeycloakKey), tenant, clientUuid)
	clientParamsBytes, err := json.Marshal(map[string]interface{}{
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
		slog.Error(commandName, "json.Marshal error", "")
		panic(err)
	}

	DoPutReturnNoContent(commandName, clientPutRequestUrl, enableDebug, clientParamsBytes, clientHeaders)
}
