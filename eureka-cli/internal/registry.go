package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func GetEurekaRegistryAuthToken(commandName string) string {
	session, err := session.NewSession()
	if err != nil {
		slog.Error(commandName, MessageKey, "session.NewSession() error")
		panic(err)
	}

	ecrClient := ecr.New(session)

	authToken, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		panic(err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		slog.Error(commandName, MessageKey, "base64.StdEncoding.DecodeString error")
		panic(err)
	}

	authCreds := strings.Split(string(decodedBytes), ":")

	jsonBytes, err := json.Marshal(map[string]string{"username": authCreds[0], "password": authCreds[1]})
	if err != nil {
		slog.Error(commandName, MessageKey, "json.Marshal error")
		panic(err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString(jsonBytes)

	slog.Info(commandName, MessageKey, "Created registry auth token")

	return encodedAuth
}

func GetModulesFromRegistries(commandName string, installJsonUrls map[string]string) map[string][]RegistryModule {
	registryModulesMap := make(map[string][]RegistryModule)

	for registryName, installJsonUrl := range installJsonUrls {
		var registryModules []RegistryModule

		installJsonResp, err := http.Get(installJsonUrl)
		if err != nil {
			slog.Error(commandName, MessageKey, "http.Get error")
			panic(err)
		}
		defer installJsonResp.Body.Close()

		err = json.NewDecoder(installJsonResp.Body).Decode(&registryModules)
		if err != nil {
			slog.Error(commandName, MessageKey, "json.NewDecoder error")
			panic(err)
		}

		if len(registryModules) > 0 {
			slog.Info(commandName, fmt.Sprintf("Found %s registry modules", registryName), len(registryModules))

			sort.Slice(registryModules, func(i, j int) bool {
				switch strings.Compare(registryModules[i].Id, registryModules[j].Id) {
				case -1:
					return true
				case 1:
					return false
				}

				return registryModules[i].Id > registryModules[j].Id
			})
		}

		registryModulesMap[registryName] = registryModules
	}

	return registryModulesMap
}
