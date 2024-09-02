package internal

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

func GetKeycloakAccessToken(commandName string, enableDebug bool, vaultRootToken string, tenant string) string {
	secretMap := GetVaultSecretKey(commandName, enableDebug, vaultRootToken, fmt.Sprintf("folio/%s", tenant))

	clientId := GetEnvironmentFromMapByKey(commandName, "KC_SERVICE_CLIENT_ID")
	clientSecret := secretMap[clientId].(string)
	systemUser := fmt.Sprintf("%s-system-user", tenant)
	systemUserPassword := secretMap[systemUser].(string)
	requestUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", viper.GetString(ResourcesKeycloakKey), tenant)

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", clientId)
	formData.Set("client_secret", clientSecret)
	formData.Set("username", systemUser)
	formData.Set("password", systemUserPassword)

	tokensMap := DoPostReturnMapStringInteface(commandName, requestUrl, enableDebug, formData, map[string]string{ContentTypeHeader: FormUrlEncodedContentType})
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "")
	}

	return tokensMap["access_token"].(string)
}
