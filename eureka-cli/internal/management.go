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
	GatewayPort = 8000

	ApplicationsPort       = 9901
	TenantsPort            = 9902
	TenantEntitlementsPort = 9903

	JsonContentType           string = "application/json"
	FormUrlEncodedContentType string = "application/x-www-form-urlencoded"

	ContentTypeHeader = "Content-Type"
	TenantHeader      = "X-Okapi-Tenant"
	TokenHeader       = "X-Okapi-Token"
)

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name        string
	SidecarName string
	Version     string
}

type RegistryModules []RegistryModule

type Applications struct {
	ApplicationDescriptors []map[string]interface{} `json:"applicationDescriptors"`
	TotalRecords           int                      `json:"totalRecords"`
}

func ExtractModuleNameAndVersion(commandName string, enableDebug bool, registryModulesMap map[string][]RegistryModule) {
	for registryName, registryModules := range registryModulesMap {
		slog.Info(commandName, fmt.Sprintf("Registering %s registry modules", registryName), "")

		for moduleIndex, module := range registryModules {
			module.Name = ModuleIdRegexp.ReplaceAllString(module.Id, `$1`)
			module.Version = ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`)
			module.Name = TrimModuleName(module.Name)
			module.SidecarName = fmt.Sprintf("%s-sc", module.Name)

			registryModules[moduleIndex] = module
		}
	}
}

func RemoveApplications(commandName string, moduleName string, enableDebug bool, panicOnError bool) {
	resp := DoGetReturnResponse(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), enableDebug, panicOnError)
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	var applications Applications

	err := json.NewDecoder(resp.Body).Decode(&applications)
	if err != nil {
		slog.Error(commandName, "json.NewDecoder error", "")
		panic(err)
	}

	for _, v := range applications.ApplicationDescriptors {
		id := v["id"].(string)

		if moduleName != "" {
			moduleNameFiltered := ModuleIdRegexp.ReplaceAllString(id, `$1`)
			if TrimModuleName(moduleNameFiltered) != moduleName {
				continue
			}
		}

		slog.Info(commandName, "Removing application", id)

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

	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)
	applicationPlatform := applicationMap["platform"].(string)
	applicationFetchDescriptors := applicationMap["fetch-descriptors"].(bool)

	if applicationMap["dependencies"] != nil {
		dependencies = applicationMap["dependencies"].([]interface{})
	}

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, fmt.Sprintf("Registering %s registry modules", registryName), "")

		for _, module := range registryModules {
			if strings.Contains(module.Name, ManagementModulePattern) {
				slog.Info(commandName, fmt.Sprintf("Ignoring %s module", module.Name), "")
				continue
			}

			_, okBackend := dto.BackendModulesMap[module.Name]
			_, okFrontend := dto.FrontendModulesMap[module.Name]
			if !okBackend && !okFrontend {
				continue
			}

			_, err := dto.FileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", TransformToEnvVar(module.Name), module.Version))
			if err != nil {
				slog.Error(commandName, "dto.FileModuleEnvPointer.WriteString error", "")
				panic(err)
			}

			moduleDescriptorUrl := fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id)

			if applicationFetchDescriptors {
				dto.ModuleDescriptorsMap[module.Id] = DoGetDecodeReturnInterface(commandName, moduleDescriptorUrl, enableDebug)
			}

			if okBackend {
				backendModule := map[string]string{"id": module.Id, "name": module.Name, "version": module.Version}
				if applicationFetchDescriptors {
					backendModuleDescriptors = append(backendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])
				} else {
					backendModule["url"] = moduleDescriptorUrl
				}

				backendModules = append(backendModules, backendModule)

				sidecarUrl := fmt.Sprintf("http://%s.eureka:%s", module.SidecarName, ServerPort)

				discoveryModules = append(discoveryModules, map[string]string{"id": module.Id, "name": module.Name, "version": module.Version, "location": sidecarUrl})
			} else if okFrontend {
				frontendModule := map[string]string{"id": module.Id, "name": module.Name, "version": module.Version}
				if applicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])
				} else {
					frontendModule["url"] = moduleDescriptorUrl
				}

				frontendModules = append(frontendModules, frontendModule)
			}

			slog.Info(commandName, "Found module", module.Name)
		}
	}

	applicationBytes, err := json.Marshal(map[string]interface{}{
		"id":                  fmt.Sprintf("%s-%s", applicationName, applicationVersion),
		"name":                applicationName,
		"version":             applicationVersion,
		"description":         "Default",
		"platform":            applicationPlatform,
		"dependencies":        dependencies,
		"modules":             backendModules,
		"uiModules":           frontendModules,
		"moduleDescriptors":   backendModuleDescriptors,
		"uiModuleDescriptors": frontendModuleDescriptors,
	})
	if err != nil {
		slog.Error(commandName, "json.Marshal error", "")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications?check=true"), enableDebug, applicationBytes, map[string]string{ContentTypeHeader: JsonContentType})

	applicationDiscoveryBytes, err := json.Marshal(map[string]interface{}{"discovery": discoveryModules})
	if err != nil {
		slog.Error(commandName, "json.Marshal error", "")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), enableDebug, applicationDiscoveryBytes, map[string]string{ContentTypeHeader: JsonContentType})
}

func GetTenants(commandName string, enableDebug bool, panicOnError bool) []interface{} {
	var foundTenants []interface{}

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), enableDebug, panicOnError)
	if foundTenantsMap["tenants"] == nil || len(foundTenantsMap["tenants"].([]interface{})) == 0 {
		return nil
	}

	foundTenants = foundTenantsMap["tenants"].([]interface{})

	return foundTenants
}

func RemoveTenants(commandName string, enableDebug bool, panicOnError bool) {
	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]interface{})
		tenantName := mapEntry["name"].(string)

		if slices.Contains(viper.GetStringSlice(TenantsKey), tenantName) {
			tenantId := mapEntry["id"].(string)

			DoDelete(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, fmt.Sprintf("/tenants/%s?purge=true", tenantId)), enableDebug)

			slog.Info(commandName, fmt.Sprintf("Removed tenant %s", tenantName), "")

			break
		}
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants")
	tenants := viper.GetStringSlice(TenantsKey)

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{"name": tenant, "description": "Default"})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantBytes, map[string]string{ContentTypeHeader: JsonContentType})

		slog.Info(commandName, fmt.Sprintf("Created %s tenant", tenant), "")
	}
}

func RemoveTenantEntitlements(commandName string, enableDebug bool, panicOnError bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]interface{})
		tenantName := mapEntry["name"].(string)

		if slices.Contains(viper.GetStringSlice(TenantsKey), tenantName) {
			tenantId := mapEntry["id"].(string)

			var applications []string
			applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

			tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
			if err != nil {
				slog.Error(commandName, "json.Marshal error", "")
				panic(err)
			}

			DoDeleteWithBody(commandName, requestUrl, enableDebug, tenantEntitlementBytes, true)

			slog.Info(commandName, fmt.Sprintf("Removed tenant entitlement %s tenant", tenantName), "")
		}
	}
}

func CreateTenantEntitlement(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=loadReference=true,loadSample=true")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenantName := mapEntry["name"].(string)

		if slices.Contains(viper.GetStringSlice(TenantsKey), tenantName) {
			tenantId := mapEntry["id"].(string)

			var applications []string
			applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

			tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
			if err != nil {
				slog.Error(commandName, "json.Marshal error", "")
				panic(err)
			}

			DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantEntitlementBytes, map[string]string{ContentTypeHeader: JsonContentType})

			slog.Info(commandName, fmt.Sprintf("Created tenant entitlement for %s tenant (%s)", tenantName, tenantId), "")
		}
	}
}

func RemoveUsers(commandName string, enableDebug bool, panicOnError bool) {

}

func CreateUsers(commandName string, enableDebug bool, accessToken string) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/users-keycloak/users")
	usersMap := viper.GetStringMap(UsersKey)

	for _, value := range usersMap {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["tenant"].(string)
		username := mapEntry["username"].(string)
		firstName := mapEntry["username"].(string)
		lastName := mapEntry["username"].(string)

		userBytes, err := json.Marshal(map[string]any{
			"username": username,
			"active":   true,
			"type":     "staff",
			"personal": map[string]any{
				"firstName":              firstName,
				"lastName":               lastName,
				"preferredContactTypeId": "002",
			},
		})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		headers := map[string]string{
			ContentTypeHeader: JsonContentType,
			TenantHeader:      tenant,
			TokenHeader:       accessToken,
		}

		DoPostReturnNoContent(DockerInternalUrl, requestUrl, enableDebug, userBytes, headers)
	}
}
