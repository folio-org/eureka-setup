package registrystep

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/viper"
)

type RegistryStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *RegistryStep {
	return &RegistryStep{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (rs *RegistryStep) ExtractModuleNameAndVersion(registryModulesMap map[string][]*models.RegistryModule, printOutput bool) {
	for registryName, registryModules := range registryModulesMap {
		if printOutput {
			slog.Info(rs.Action.Name, "text", fmt.Sprintf("Extracting %s registry module names and versions", registryName))
		}

		for moduleIndex, module := range registryModules {
			if module.Id == "okapi" {
				continue
			}

			module.Name = helpers.GetModuleNameFromID(module.Id)
			module.Version = helpers.GetModuleVersionPFromID(module.Id)

			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = fmt.Sprintf("%s-sc", module.Name)
			}

			registryModules[moduleIndex] = module
		}
	}
}

func (rs *RegistryStep) GetAuthTokenIfPresent() (string, error) {
	if os.Getenv(constant.ECRRepositoryEnv) == "" {
		return "", nil
	}

	session, err := session.NewSession()
	if err != nil {
		return "", err
	}

	authToken, err := ecr.New(session).GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", err
	}

	authCreds := strings.Split(string(decodedBytes), ":")

	jsonBytes, err := json.Marshal(map[string]string{"username": authCreds[0], "password": authCreds[1]})
	if err != nil {
		return "", err
	}

	encodedAuth := base64.StdEncoding.EncodeToString(jsonBytes)

	slog.Error(rs.Action.Name, "error", err)

	return encodedAuth, nil
}

func (rs *RegistryStep) GetModules(installJsonURLs map[string]string, printOutput bool) (map[string][]*models.RegistryModule, error) {
	registryModulesMap := make(map[string][]*models.RegistryModule)

	for registryName, installJsonURL := range installJsonURLs {
		installJsonResp, err := rs.HTTPClient.GetReturnResponse(installJsonURL, map[string]string{})
		if err != nil {
			return nil, err
		}
		defer helpers.CloseReader(installJsonResp.Body)

		var registryModules []*models.RegistryModule
		err = json.NewDecoder(installJsonResp.Body).Decode(&registryModules)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if registryName == constant.FolioRegistry {
			for name, value := range viper.GetStringMap(field.CustomFrontendModules) {
				if value == nil {
					continue
				}

				mapEntry := value.(map[string]any)
				if mapEntry[field.ModuleVersionEntry] == nil {
					continue
				}

				registryModule := &models.RegistryModule{Id: fmt.Sprintf("%s-%s", name, mapEntry[field.ModuleVersionEntry].(string)), Action: "enable"}

				registryModules = append(registryModules, registryModule)
			}
		}

		if len(registryModules) > 0 {
			if printOutput {
				slog.Info(rs.Action.Name, "text", fmt.Sprintf("Read %s registry with %d modules", registryName, len(registryModules)))
			}

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

	return registryModulesMap, nil
}

func (rs *RegistryStep) GetNamespace(version string) string {
	namespace := os.Getenv(constant.ECRRepositoryEnv)
	if namespace != "" {
		slog.Info(rs.Action.Name, "text", fmt.Sprintf("Using AWS ECR registry namespace: %s", namespace))

		return namespace
	}

	if strings.Contains(version, "SNAPSHOT") {
		namespace = constant.SnapshotRegistry
	} else {
		namespace = constant.ReleaseRegistry
	}

	return namespace
}
