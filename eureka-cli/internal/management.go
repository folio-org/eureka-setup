package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

const (
	ApplicationsPort       = 9901
	TenantsPort            = 9902
	TenantEntitlementsPort = 9903
)

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name    string
	Version string
}

type RegistryModules []RegistryModule

type Applications struct {
	ApplicationDescriptors []map[string]interface{} `json:"applicationDescriptors"`
	TotalRecords           int                      `json:"totalRecords"`
}

func ExtractModuleNameAndVersion(commandName string, enableDebug bool, registryModulesMap map[string][]RegistryModule) {
	for registryName, registryModules := range registryModulesMap {
		slog.Info(commandName, MessageKey, fmt.Sprintf("Registering %s registry modules", registryName))

		for moduleIndex, module := range registryModules {
			module.Name = ModuleIdRegexp.ReplaceAllString(module.Id, `$1`)
			module.Version = ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`)
			module.Name = TrimModuleName(module.Name)

			registryModules[moduleIndex] = module
		}
	}
}

func DeregisterModules(commandName string, moduleName string, enableDebug bool) {
	slog.Info(commandName, "Deregistering module", moduleName)

	resp := DoGetReturnResponse(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), enableDebug)
	defer resp.Body.Close()

	var apps Applications

	err := json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		slog.Error(commandName, MessageKey, "json.NewDecoder error")
		panic(err)
	}

	if apps.TotalRecords == 0 {
		slog.Info(commandName, MessageKey, "No deployed module applications were found")
	}

	if apps.TotalRecords > 0 {
		for _, v := range apps.ApplicationDescriptors {
			id := v["id"].(string)

			if moduleName != "" {
				moduleNameFiltered := ModuleIdRegexp.ReplaceAllString(id, `$1`)
				moduleNameFiltered = TrimModuleName(moduleNameFiltered)

				if moduleNameFiltered != moduleName {
					continue
				}
			}

			slog.Info(commandName, "Deregistering application", id)

			DoDelete(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, fmt.Sprintf("/applications/%s", id)), enableDebug)
		}
	}
}

func RegisterModules(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, MessageKey, fmt.Sprintf("Registering %s registry modules", registryName))

		for _, module := range registryModules {
			if strings.Contains(module.Name, ManagementModulePattern) {
				slog.Info(commandName, MessageKey, fmt.Sprintf("Ignoring %s module", module.Name))

				continue
			}

			_, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			moduleNameEnv := EnvNameRegexp.ReplaceAllString(strings.ToUpper(module.Name), `_`)

			_, err := dto.FileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", moduleNameEnv, module.Version))
			if err != nil {
				slog.Error(commandName, MessageKey, "moduleEnvVarsFile.WriteString error")
				panic(err)
			}

			moduleDescriptors := DoGetDecodeReturnInterface(commandName, fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id), enableDebug)

			dto.ModuleDescriptorsMap[module.Id] = moduleDescriptors

			var appModules []map[string]string
			appModules = append(appModules, map[string]string{"name": module.Name, "version": module.Version})

			appBytes, err := json.Marshal(map[string]interface{}{
				"id":                module.Id,
				"version":           module.Version,
				"name":              module.Name,
				"description":       "Deployed by Eureka CLI",
				"modules":           appModules,
				"moduleDescriptors": moduleDescriptors,
			})
			if err != nil {
				slog.Error(commandName, MessageKey, "json.Marshal error")
				panic(err)
			}

			DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), enableDebug, appBytes)

			var appDiscModules []map[string]string
			appDiscModules = append(appDiscModules, map[string]string{
				"id":       module.Id,
				"name":     module.Name,
				"version":  module.Version,
				"location": fmt.Sprintf("http://%s.eureka:%s", module.Name, ServerPort),
			})

			appDiscInfoBytes, err := json.Marshal(map[string]interface{}{"discovery": appDiscModules})
			if err != nil {
				slog.Error(commandName, MessageKey, "json.Marshal error")
				panic(err)
			}

			DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), enableDebug, appDiscInfoBytes)

			slog.Info(commandName, "Registered module", module.Id)
		}
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	tenants := viper.GetStringSlice(TenantConfigKey)

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug)

	foundTenants := foundTenantsMap["tenants"].([]interface{})

	for _, value := range foundTenants {
		mapEntry := value.(map[string]interface{})

		name := mapEntry["name"].(string)

		if slices.Contains(tenants, name) {
			id := mapEntry["id"].(string)

			DoDelete(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, fmt.Sprintf("/tenants/%s?purge=true", id)), enableDebug)

			slog.Info(commandName, MessageKey, fmt.Sprintf("Removed tenant %s", name))

			break
		}
	}

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{
			"name":        tenant,
			"description": "Default_tenant",
		})
		if err != nil {
			slog.Error(commandName, MessageKey, "json.Marshal error")
			panic(err)
		}

		DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug, tenantBytes)

		slog.Info(commandName, MessageKey, fmt.Sprintf("Created tenant %s", tenant))
	}
}
