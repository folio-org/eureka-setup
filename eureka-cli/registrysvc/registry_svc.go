package registrysvc

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

type RegistrySvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *RegistrySvc {
	return &RegistrySvc{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (rs *RegistrySvc) ExtractModuleNameAndVersion(rr1 map[string][]*models.RegistryModule) {
	for _, rr2 := range rr1 {
		for moduleIndex, module := range rr2 {
			if module.ID == "okapi" {
				continue
			}

			module.Name = helpers.GetModuleNameFromID(module.ID)
			module.Version = helpers.GetModuleVersionPFromID(module.ID)

			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = fmt.Sprintf("%s-sc", module.Name)
			}

			rr2[moduleIndex] = module
		}
	}
}

func (rs *RegistrySvc) GetAuthTokenIfPresent() (string, error) {
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

	bb, err := json.Marshal(map[string]string{
		"username": authCreds[0],
		"password": authCreds[1],
	})
	if err != nil {
		return "", err
	}

	encodedAuth := base64.StdEncoding.EncodeToString(bb)

	slog.Error(rs.Action.Name, "error", err)

	return encodedAuth, nil
}

func (rs *RegistrySvc) GetModules(installJsonURLs map[string]string, printOutput bool) (map[string][]*models.RegistryModule, error) {
	rr1 := make(map[string][]*models.RegistryModule)

	for registryName, installJsonURL := range installJsonURLs {
		installJsonResp, err := rs.HTTPClient.GetReturnResponse(installJsonURL, map[string]string{})
		if err != nil {
			return nil, err
		}
		defer helpers.CloseReader(installJsonResp.Body)

		var rr2 []*models.RegistryModule
		err = json.NewDecoder(installJsonResp.Body).Decode(&rr2)
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

				registryModule := &models.RegistryModule{
					ID:     fmt.Sprintf("%s-%s", name, mapEntry[field.ModuleVersionEntry].(string)),
					Action: "enable",
				}

				rr2 = append(rr2, registryModule)
			}
		}

		if len(rr2) > 0 {
			if printOutput {
				slog.Info(rs.Action.Name, "text", "Read registry with modules", "registry", registryName, "moduleCount", len(rr2))
			}

			sort.Slice(rr2, func(i, j int) bool {
				switch strings.Compare(rr2[i].ID, rr2[j].ID) {
				case -1:
					return true
				case 1:
					return false
				}

				return rr2[i].ID > rr2[j].ID
			})
		}

		rr1[registryName] = rr2
	}

	return rr1, nil
}

func (rs *RegistrySvc) GetNamespace(version string) string {
	namespace := os.Getenv(constant.ECRRepositoryEnv)
	if namespace != "" {
		slog.Info(rs.Action.Name, "text", "Using AWS ECR registry namespace", "namespace", namespace)

		return namespace
	}

	if strings.Contains(version, "SNAPSHOT") {
		namespace = constant.SnapshotRegistry
	} else {
		namespace = constant.ReleaseRegistry
	}

	return namespace
}
