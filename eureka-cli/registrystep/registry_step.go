package registrystep

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

			module.Name = helpers.GetModuleName(module.Id)
			module.Version = helpers.GetModuleVersionP(module.Id)

			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = fmt.Sprintf("%s-sc", module.Name)
			}

			registryModules[moduleIndex] = module
		}
	}
}

func (rs *RegistryStep) GetAuthTokenIfPresent() string {
	// If this env variable isn't set, then assume it is a public repository and no auth token is needed.
	if os.Getenv(constant.ECRRepository) == "" {
		return ""
	}

	session, err := session.NewSession()
	if err != nil {
		slog.Error(rs.Action.Name, "error", err)
		panic(err)
	}

	authToken, err := ecr.New(session).GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		panic(err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		slog.Error(rs.Action.Name, "error", err)
		panic(err)
	}

	authCreds := strings.Split(string(decodedBytes), ":")

	jsonBytes, err := json.Marshal(map[string]string{"username": authCreds[0], "password": authCreds[1]})
	if err != nil {
		slog.Error(rs.Action.Name, "error", err)
		panic(err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString(jsonBytes)

	slog.Error(rs.Action.Name, "error", err)

	return encodedAuth
}

func (rs *RegistryStep) GetModules(installJsonURLs map[string]string, printOutput bool) map[string][]*models.RegistryModule {
	registryModulesMap := make(map[string][]*models.RegistryModule)

	for registryName, installJsonURL := range installJsonURLs {
		var registryModules []*models.RegistryModule

		installJsonResp, err := http.Get(installJsonURL)
		if err != nil {
			slog.Error(rs.Action.Name, "error", err)
			panic(err)
		}
		defer func() {
			_ = installJsonResp.Body.Close()
		}()

		err = json.NewDecoder(installJsonResp.Body).Decode(&registryModules)
		if err != nil {
			slog.Error(rs.Action.Name, "error", err)
			panic(err)
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

				registryModules = append(registryModules, &models.RegistryModule{Id: fmt.Sprintf("%s-%s", name, mapEntry[field.ModuleVersionEntry].(string)), Action: "enable"})
			}
		}

		if len(registryModules) > 0 {
			if printOutput {
				slog.Info(rs.Action.Name, "text", fmt.Sprintf("Found %s modules: %d", registryName, len(registryModules)))
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

	return registryModulesMap
}

func (rs *RegistryStep) GetNamespace(version string) string {
	// AWS ECR Folio registry should be considered a secret because it has an account id in it so we put it in the env.
	namespace := os.Getenv(constant.ECRRepository)
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
