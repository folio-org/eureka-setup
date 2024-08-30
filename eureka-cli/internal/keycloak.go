package internal

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

type DirectGrantTokenDto struct {
	realmName    string
	clientId     string
	clientSecret string
	username     string
	password     string
}

func NewDirectGrantTokenDto(realmName string, clientId string, clientSecret string, username string, password string) *DirectGrantTokenDto {
	return &DirectGrantTokenDto{
		realmName:    realmName,
		clientId:     clientId,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
	}
}

func GetKeycloakAccessToken(commandName string, enableDebug bool, dto *DirectGrantTokenDto) string {
	requestUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", viper.GetString(ResourcesKeycloakKey), dto.realmName)

	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("client_id", dto.clientId)
	formData.Set("client_secret", dto.clientSecret)
	formData.Set("username", dto.username)
	formData.Set("password", dto.password)

	tokensMap := DoPostReturnMapStringInteface(commandName, requestUrl, enableDebug, formData, map[string]string{ContentTypeHeader: FormUrlEncodedContentType})
	if tokensMap["access_token"] == nil {
		LogErrorPanic(commandName, "")
	}

	return tokensMap["access_token"].(string)
}
