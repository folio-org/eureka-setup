package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/viper"
)

const (
	SnapshotRegistry string = "folioci"
	ReleaseRegistry  string = "folioorg"

	awsEcrFolioRepoEnvKey string = "AWS_ECR_FOLIO_REPO"
)

func GetRegistryAuthTokenIfPresent(commandName string) string {
	// If this env variable isn't set, then assume it is a public repository and no auth token is needed.
	if os.Getenv(awsEcrFolioRepoEnvKey) == "" {
		return ""
	}

	session, err := session.NewSession()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "session.NewSession() error")
		panic(err)
	}

	authToken, err := ecr.New(session).GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		panic(err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "base64.StdEncoding.DecodeString error")
		panic(err)
	}

	authCreds := strings.Split(string(decodedBytes), ":")

	jsonBytes, err := json.Marshal(map[string]string{"username": authCreds[0], "password": authCreds[1]})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString(jsonBytes)

	slog.Info(commandName, GetFuncName(), "Created registry auth token")

	return encodedAuth
}

func GetModulesFromRegistries(commandName string, installJsonUrls map[string]string) map[string][]*RegistryModule {
	registryModulesMap := make(map[string][]*RegistryModule)

	for registryName, installJsonUrl := range installJsonUrls {
		var registryModules []*RegistryModule

		installJsonResp, err := http.Get(installJsonUrl)
		if err != nil {
			slog.Error(commandName, GetFuncName(), "http.Get error")
			panic(err)
		}
		defer installJsonResp.Body.Close()

		err = json.NewDecoder(installJsonResp.Body).Decode(&registryModules)
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
			panic(err)
		}

		if registryName == FolioRegistry {
			for name, value := range viper.GetStringMap(CustomFrontendModuleKey) {
				if value == nil {
					continue
				}

				mapEntry := value.(map[string]interface{})
				if mapEntry["version"] == nil {
					continue
				}

				registryModules = append(registryModules, &RegistryModule{Id: fmt.Sprintf("%s-%s", name, mapEntry["version"].(string)), Action: "enable"})
			}
		}

		if len(registryModules) > 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found %s modules: %d", registryName, len(registryModules)))

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

func GetImageRegistryNamespace(commandName string, version string) string {
	var registryNamespace string
	// AWS ECR Folio registry should be considered a secret because it has an account id in it so we put it in the env.
	registryNamespace = os.Getenv(awsEcrFolioRepoEnvKey)

	if registryNamespace != "" {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Using AWS ECR registry namespace: %s", registryNamespace))

		return registryNamespace
	}

	if strings.Contains(version, "SNAPSHOT") {
		registryNamespace = SnapshotRegistry
	} else {
		registryNamespace = ReleaseRegistry
	}

	return registryNamespace
}
