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

func RemoveApplications(commandName string, moduleName string, enableDebug bool) {
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

func CreateApplications(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []interface{}
		frontendModuleDescriptors []interface{}
		dependencies              []interface{}
		discoveryModules          []map[string]string
	)

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, MessageKey, fmt.Sprintf("Registering %s registry modules", registryName))

		for _, module := range registryModules {
			if strings.Contains(module.Name, ManagementModulePattern) {
				slog.Info(commandName, MessageKey, fmt.Sprintf("Ignoring %s module", module.Name))
				continue
			}

			_, okBackend := dto.BackendModulesMap[module.Name]
			_, okFrontend := dto.FrontendModulesMap[module.Name]
			if !okBackend && !okFrontend {
				continue
			}

			_, err := dto.FileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", TransformToEnvVar(module.Name), module.Version))
			if err != nil {
				slog.Error(commandName, MessageKey, "moduleEnvVarsFile.WriteString error")
				panic(err)
			}

			url := fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id)

			dto.ModuleDescriptorsMap[module.Id] = DoGetDecodeReturnInterface(commandName, url, enableDebug)

			if okBackend {
				backendModules = append(backendModules, map[string]string{
					"id":      module.Id,
					"name":    module.Name,
					"version": module.Version,
				})
				backendModuleDescriptors = append(backendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])

				discoveryModules = append(discoveryModules, map[string]string{
					"id":       module.Id,
					"name":     module.Name,
					"version":  module.Version,
					"location": fmt.Sprintf("http://%s.eureka:%s", module.Name, ServerPort),
				})
			} else if okFrontend {
				frontendModules = append(frontendModules, map[string]string{
					"id":      module.Id,
					"name":    module.Name,
					"version": module.Version,
				})
				frontendModuleDescriptors = append(frontendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])
			}

			slog.Info(commandName, "Found module", module.Name)
		}
	}

	// ########## Application ##########
	appBytes, err := json.Marshal(map[string]interface{}{
		"id":                  "app-platform-complete-1.0.0",
		"name":                "app-platform-complete",
		"version":             "1.0.0",
		"description":         "app-platform-complete:Deployed by Eureka CLI",
		"platform":            "base",
		"dependencies":        dependencies,
		"modules":             backendModules,
		"uiModules":           frontendModules,
		"moduleDescriptors":   backendModuleDescriptors,
		"uiModuleDescriptors": frontendModuleDescriptors,
	})
	if err != nil {
		slog.Error(commandName, MessageKey, "json.Marshal error")
		panic(err)
	}

	DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications?check=false"), enableDebug, appBytes)

	// ########## Application Discovery ##########
	applicationDiscoveryBytes, err := json.Marshal(map[string]interface{}{"discovery": discoveryModules})
	if err != nil {
		slog.Error(commandName, MessageKey, "json.Marshal error")
		panic(err)
	}

	DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), enableDebug, applicationDiscoveryBytes)
}

func RemoveTenants(commandName string, enableDebug bool) {
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
		tenantBytes, err := json.Marshal(map[string]string{"name": tenant, "description": "Default_tenant"})
		if err != nil {
			slog.Error(commandName, MessageKey, "json.Marshal error")
			panic(err)
		}

		DoPostNoContent(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug, tenantBytes)

		slog.Info(commandName, MessageKey, fmt.Sprintf("Created tenant named %s", tenant))
	}
}

func RemoveTenantEntitlements(commandName string, enableDebug bool) {
	tenants := viper.GetStringSlice(TenantConfigKey)

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug)

	foundTenants := foundTenantsMap["tenants"].([]interface{})

	for _, value := range foundTenants {
		mapEntry := value.(map[string]interface{})

		name := mapEntry["name"].(string)

		if slices.Contains(tenants, name) {
			id := mapEntry["id"].(string)

			var applications []string
			applications = append(applications, "app-platform-complete-1.0.0")

			tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": id, "applications": applications})
			if err != nil {
				slog.Error(commandName, MessageKey, "json.Marshal error")
				panic(err)
			}

			url := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=loadReference=false,loadSample=false&purge=true")

			DoDeleteBody(commandName, url, enableDebug, tenantEntitlementBytes)

			slog.Info(commandName, MessageKey, fmt.Sprintf("Removed tenant entitlement for tenant name %s and id %s", name, id))
		}
	}
}

func CreateTenantEntitlement(commandName string, enableDebug bool) {
	tenants := viper.GetStringSlice(TenantConfigKey)

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug)

	foundTenants := foundTenantsMap["tenants"].([]interface{})

	for _, value := range foundTenants {
		mapEntry := value.(map[string]interface{})

		name := mapEntry["name"].(string)

		if slices.Contains(tenants, name) {
			id := mapEntry["id"].(string)

			var applications []string
			applications = append(applications, "app-platform-complete-1.0.0")

			tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": id, "applications": applications})
			if err != nil {
				slog.Error(commandName, MessageKey, "json.Marshal error")
				panic(err)
			}

			url := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=true&tenantParameters=loadReference=false,loadSample=false&includeModules=true")

			DoPostNoContent(commandName, url, enableDebug, tenantEntitlementBytes)

			slog.Info(commandName, MessageKey, fmt.Sprintf("Created tenant entitlement for tenant name %s and id %s", name, id))
		}
	}
}
