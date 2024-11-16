package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	GatewayPort int = 8000

	ApplicationsPort       int = 9901
	TenantsPort            int = 9902
	TenantEntitlementsPort int = 9903

	JsonContentType           string = "application/json"
	FormUrlEncodedContentType string = "application/x-www-form-urlencoded"

	ContentTypeHeader   string = "Content-Type"
	AuthorizationHeader string = "Authorization"
	TenantHeader        string = "X-Okapi-Tenant"
	TokenHeader         string = "X-Okapi-Token"

	RemoveRoleUnsupported bool = true

	HealtcheckUri         string = "/admin/health"
	HealtcheckMaxAttempts int    = 50
)

var HealthcheckDefaultDuration time.Duration = 30 * time.Second

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name        string
	SidecarName string
	Version     *string
}

type RegistryModules []RegistryModule

// ######## Auxiliary ########

func ExtractModuleNameAndVersion(commandName string, enableDebug bool, registryModulesMap map[string][]*RegistryModule) {
	for registryName, registryModules := range registryModulesMap {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Registering %s registry modules", registryName))

		for moduleIndex, module := range registryModules {
			module.Name = TrimModuleName(ModuleIdRegexp.ReplaceAllString(module.Id, `$1`))
			moduleVersion := ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`)
			module.Version = &moduleVersion
			module.SidecarName = fmt.Sprintf("%s-sc", module.Name)

			registryModules[moduleIndex] = module
		}
	}
}

func PerformModuleHealthcheck(commandName string, enableDebug bool, waitMutex *sync.WaitGroup, moduleName string, port int) {
	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Waiting for module container %s on port %d to initialize", moduleName, port))
	requestUrl := fmt.Sprintf(DockerInternalUrl, port, HealtcheckUri)
	healthcheckAttempts := HealtcheckMaxAttempts
	for {
		time.Sleep(HealthcheckDefaultDuration)

		isHealthyVertxContainer := false
		actuatorHealthStr := DoGetDecodeReturnString(commandName, requestUrl, enableDebug, false, map[string]string{})
		if strings.Contains(actuatorHealthStr, "OK") {
			isHealthyVertxContainer = !isHealthyVertxContainer
		}

		isHealthySpringBootContainer := false
		actuatorHealthMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, false, map[string]string{})
		if actuatorHealthMap != nil && strings.Contains(actuatorHealthMap["status"].(string), "UP") {
			isHealthySpringBootContainer = !isHealthySpringBootContainer
		}

		if isHealthyVertxContainer || isHealthySpringBootContainer {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is healthy", moduleName))
			waitMutex.Done()
			break
		}

		healthcheckAttempts--
		if healthcheckAttempts > 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is unhealthy, %d/%d attempts left", moduleName, healthcheckAttempts, HealtcheckMaxAttempts))
		} else {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is unhealthy, out of attempts", moduleName))
			waitMutex.Done()
			LogErrorPanic(commandName, fmt.Sprintf("internal.PerformModuleHealthcheck - Module container %s did not initialize, cannot continue", moduleName))
		}
	}
}

// ######## Application & Application Discovery ########

func RemoveApplications(commandName string, moduleName string, enableDebug bool, panicOnError bool) {
	resp := DoGetReturnResponse(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), enableDebug, panicOnError, map[string]string{})
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	type Applications struct {
		ApplicationDescriptors []map[string]interface{} `json:"applicationDescriptors"`
		TotalRecords           int                      `json:"totalRecords"`
	}

	var applications Applications

	err := json.NewDecoder(resp.Body).Decode(&applications)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
		panic(err)
	}

	for _, value := range applications.ApplicationDescriptors {
		id := value["id"].(string)

		if moduleName != "" {
			moduleNameFiltered := ModuleIdRegexp.ReplaceAllString(id, `$1`)
			if TrimModuleName(moduleNameFiltered) != moduleName {
				continue
			}
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, ApplicationsPort, fmt.Sprintf("/applications/%s", id))

		DoDelete(commandName, requestUrl, enableDebug, false, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Removed '%s' application`, id))
	}
}

func CreateApplications(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []interface{}
		frontendModuleDescriptors []interface{}
		dependencies              map[string]interface{}
		discoveryModules          []map[string]string
	)

	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)
	applicationPlatform := applicationMap["platform"].(string)
	applicationFetchDescriptors := applicationMap["fetch-descriptors"].(bool)

	if applicationMap["dependencies"] != nil {
		dependencies = applicationMap["dependencies"].(map[string]interface{})
	}

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Registering %s modules", registryName))

		for _, module := range registryModules {
			if strings.Contains(module.Name, ManagementModulePattern) {
				slog.Info(commandName, GetFuncName(), fmt.Sprintf("Ignoring %s management module registration", module.Name))
				continue
			}

			backendModule, okBackend := dto.BackendModulesMap[module.Name]

			frontendModule, okFrontend := dto.FrontendModulesMap[module.Name]
			if (!okBackend && !okFrontend) || (okBackend && !backendModule.DeployModule || okFrontend && !frontendModule.DeployModule) {
				continue
			}

			if okBackend && backendModule.ModuleVersion != nil || okFrontend && frontendModule.ModuleVersion != nil {
				if backendModule.ModuleVersion != nil {
					module.Version = backendModule.ModuleVersion
				} else if frontendModule.ModuleVersion != nil {
					module.Version = frontendModule.ModuleVersion
				}
				module.Id = fmt.Sprintf("%s-%s", module.Name, *module.Version)
			}

			_, err := dto.FileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", TransformToEnvVar(module.Name), *module.Version))
			if err != nil {
				slog.Error(commandName, GetFuncName(), "dto.FileModuleEnvPointer.WriteString error")
				panic(err)
			}

			moduleDescriptorUrl := fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id)

			if applicationFetchDescriptors {
				dto.ModuleDescriptorsMap[module.Id] = DoGetDecodeReturnInterface(commandName, moduleDescriptorUrl, enableDebug, true, map[string]string{})
			}

			if okBackend {
				serverPort := strconv.Itoa(backendModule.ModuleServerPort)
				backendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}

				if applicationFetchDescriptors {
					backendModuleDescriptors = append(backendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])
				} else {
					backendModule["url"] = moduleDescriptorUrl
				}

				backendModules = append(backendModules, backendModule)

				sidecarUrl := fmt.Sprintf("http://%s.eureka:%s", module.SidecarName, serverPort)

				discoveryModules = append(discoveryModules, map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version, "location": sidecarUrl})
			} else if okFrontend {
				frontendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}
				if applicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, dto.ModuleDescriptorsMap[module.Id])
				} else {
					frontendModule["url"] = moduleDescriptorUrl
				}

				frontendModules = append(frontendModules, frontendModule)
			}

			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found module for registration: %s with version %s", module.Name, *module.Version))
		}
	}

	applicationId := fmt.Sprintf("%s-%s", applicationName, applicationVersion)

	applicationBytes, err := json.Marshal(map[string]interface{}{
		"id":                  applicationId,
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
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications?check=true"), enableDebug, applicationBytes, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created '%s' application`, applicationId))

	applicationDiscoveryBytes, err := json.Marshal(map[string]interface{}{"discovery": discoveryModules})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), enableDebug, applicationDiscoveryBytes, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created '%d' entries of application module discovery`, len(discoveryModules)))
}

func UpdateApplicationModuleDiscovery(commandName string, enableDebug bool, id string, location string, restore bool, portServer string) {
	name := TrimModuleName(ModuleIdRegexp.ReplaceAllString(id, `$1`))
	version := ModuleIdRegexp.ReplaceAllString(id, `$2$3`)
	if location == "" || restore {
		location = fmt.Sprintf("http://%s.eureka:%s", name, portServer)
	}

	applicationDiscoveryBytes, err := json.Marshal(map[string]interface{}{"id": id, "name": name, "version": version, "location": location})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	requestUrl := fmt.Sprintf(DockerInternalUrl, ApplicationsPort, fmt.Sprintf("/modules/%s/discovery", id))

	DoPutReturnNoContent(commandName, requestUrl, enableDebug, applicationDiscoveryBytes, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Updated application module discovery for '%s' module with '%s' location`, name, location))
}

// ######## Tenants ########

func GetTenants(commandName string, enableDebug bool, panicOnError bool) []interface{} {
	var foundTenants []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants")

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, map[string]string{})
	if foundTenantsMap["tenants"] == nil || len(foundTenantsMap["tenants"].([]interface{})) == 0 {
		return nil
	}

	foundTenants = foundTenantsMap["tenants"].([]interface{})

	return foundTenants
}

func RemoveTenants(commandName string, enableDebug bool, panicOnError bool) {
	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant) {
			continue
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, TenantsPort, fmt.Sprintf("/tenants/%s?purge=true", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Removed '%s' tenant`, tenant))
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants")
	tenants := ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey))

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{"name": tenant, "description": "Default"})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantBytes, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created '%s' tenant`, tenant))
	}
}

// ######## Tenant Entitlements ########

func RemoveTenantEntitlements(commandName string, enableDebug bool, panicOnError bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant) {
			continue
		}

		tenantId := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoDeleteWithBody(commandName, requestUrl, enableDebug, tenantEntitlementBytes, true, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Removed tenant entitlement for '%s' tenant`, tenant))

	}
}

// TODO Add depCheck=false flag
func CreateTenantEntitlement(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=loadReference=false,loadSample=false")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant) {
			continue
		}

		tenantId := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantEntitlementBytes, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created tenant entitlement for '%s' tenant`, tenant))

	}
}

// ######## Users ########

func GetUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) []interface{} {
	var foundUsers []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/users")
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundTenantsMap["users"] == nil || len(foundTenantsMap["users"].([]interface{})) == 0 {
		return nil
	}

	foundUsers = foundTenantsMap["users"].([]interface{})

	return foundUsers
}

func RemoveUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, value := range GetUsers(commandName, enableDebug, panicOnError, tenant, accessToken) {
		mapEntry := value.(map[string]interface{})
		username := mapEntry["username"].(string)
		usersMap := viper.GetStringMap(UsersKey)
		if usersMap[username] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Removed '%s' in '%s' realm`, username, tenant))
	}
}

func CreateUsers(commandName string, enableDebug bool, accessToken string) {
	postUserRequestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/users-keycloak/users")
	postUserPasswordRequestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/authn/credentials")
	postUserRoleRequestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles/users")
	usersMap := viper.GetStringMap(UsersKey)

	for username, value := range usersMap {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["tenant"].(string)
		password := mapEntry["password"].(string)
		firstName := mapEntry["first-name"].(string)
		lastName := mapEntry["last-name"].(string)
		userRoles := mapEntry["roles"].([]interface{})

		userBytes, err := json.Marshal(map[string]any{
			"username": username,
			"active":   true,
			"type":     "staff",
			"personal": map[string]any{
				"firstName":              firstName,
				"lastName":               lastName,
				"email":                  fmt.Sprintf("%s-%s", tenant, username),
				"preferredContactTypeId": "002",
			},
		})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		okapiBasedHeaders := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}
		nonOkapiBasedHeaders := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken)}

		createdUserMap := DoPostReturnMapStringInteface(commandName, postUserRequestUrl, enableDebug, userBytes, okapiBasedHeaders)

		userId := createdUserMap["id"].(string)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created '%s' user with password '%s' in '%s' realm`, username, password, tenant))

		userPasswordBytes, err := json.Marshal(map[string]any{"userId": userId, "username": username, "password": password})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, postUserPasswordRequestUrl, enableDebug, userPasswordBytes, nonOkapiBasedHeaders)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Attached '%s' password to '%s' user in '%s' realm`, password, username, tenant))

		var roleIds []string
		for _, userRole := range userRoles {
			role := GetRoleByName(commandName, enableDebug, userRole.(string), okapiBasedHeaders)
			roleId := role["id"].(string)
			roleName := role["name"].(string)

			if roleId == "" {
				slog.Warn(commandName, GetFuncName(), fmt.Sprintf("internal.GetRoleByName warn - Did not find role %s by name", roleName))
				continue
			}

			roleIds = append(roleIds, roleId)
		}

		userRoleBytes, err := json.Marshal(map[string]any{"userId": userId, "roleIds": roleIds})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, postUserRoleRequestUrl, enableDebug, userRoleBytes, okapiBasedHeaders)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Attached '%d' roles to '%s' user in '%s' realm`, len(roleIds), username, tenant))
	}
}

// ######## Roles ########

func GetRoles(commandName string, enableDebug bool, panicOnError bool, headers map[string]string) []interface{} {
	var foundRoles []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles")

	foundRolesMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundRolesMap["roles"] == nil || len(foundRolesMap["roles"].([]interface{})) == 0 {
		return nil
	}

	foundRoles = foundRolesMap["roles"].([]interface{})

	return foundRoles
}

func GetRoleByName(commandName string, enableDebug bool, roleName string, headers map[string]string) map[string]interface{} {
	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/roles?query=name==%s", roleName))

	foundRolesMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, true, headers)
	if foundRolesMap["roles"] == nil {
		return nil
	}

	foundRoles := foundRolesMap["roles"].([]interface{})
	if len(foundRoles) != 1 {
		LogErrorPanic(commandName, fmt.Sprintf("internal.GetRoleByName - Number of found roles by %s role name is not 1", roleName))
	}

	return foundRoles[0].(map[string]interface{})
}

func RemoveRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, value := range GetRoles(commandName, enableDebug, panicOnError, headers) {
		mapEntry := value.(map[string]interface{})
		roleName := mapEntry["name"].(string)

		rolesMap := viper.GetStringMap(RolesKey)
		if rolesMap[roleName] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/roles-keycloak/roles/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Removed '%s' role in '%s' realm`, roleName, tenant))
	}
}

func CreateRoles(commandName string, enableDebug bool, accessToken string) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles")
	rolesMap := viper.GetStringMap(RolesKey)

	for role, value := range rolesMap {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["tenant"].(string)
		caser := cases.Title(language.English)
		headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

		roleBytes, err := json.Marshal(map[string]string{"name": caser.String(role), "description": "Default"})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, roleBytes, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Created '%s' role in '%s' realm`, role, tenant))
	}
}

// ######## Capabilities ########

func GetCapabilitySets(commandName string, enableDebug bool, panicOnError bool, headers map[string]string) []interface{} {
	var foundCapabilitySets []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/capability-sets?offset=0&limit=1000")

	foundCapabilitySetsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]interface{})) == 0 {
		return nil
	}

	foundCapabilitySets = foundCapabilitySetsMap["capabilitySets"].([]interface{})

	return foundCapabilitySets
}

func GetCapabilitySetsByName(commandName string, enableDebug bool, panicOnError bool, capabilitySetName string, headers map[string]string) []interface{} {
	var foundCapabilitySets []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilitySetName))

	foundCapabilitySetsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]interface{})) == 0 {
		return nil
	}

	foundCapabilitySets = foundCapabilitySetsMap["capabilitySets"].([]interface{})

	return foundCapabilitySets
}

func DetachCapabilitySetsFromRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, value := range GetRoles(commandName, enableDebug, panicOnError, headers) {
		mapEntry := value.(map[string]interface{})
		roleName := mapEntry["name"].(string)

		rolesMap := viper.GetStringMap(RolesKey)
		if rolesMap[strings.ToLower(roleName)] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Detached capability sets from '%s' role in '%s' realm`, roleName, tenant))
	}
}

func AttachCapabilitySetsToRoles(commandName string, enableDebug bool, tenant string, accessToken string) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles/capability-sets")
	rolesMapConfig := viper.GetStringMap(RolesKey)
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, roleValue := range GetRoles(commandName, enableDebug, true, headers) {
		roleMapEntry := roleValue.(map[string]interface{})
		roleId := roleMapEntry["id"].(string)
		roleName := roleMapEntry["name"].(string)
		rolesMapConfigByRole, ok := rolesMapConfig[strings.ToLower(roleName)]
		if !ok {
			continue
		}

		roleConfigMapEntry := rolesMapConfigByRole.(map[string]interface{})
		if tenant != roleConfigMapEntry["tenant"].(string) {
			continue
		}

		capabilitySetsConfig := roleConfigMapEntry["capability-sets"].([]interface{})

		var capabilitySetsMapList []map[string]interface{}
		if len(capabilitySetsConfig) == 1 && slices.Contains(capabilitySetsConfig, "all") {
			for _, capabilityValue := range GetCapabilitySets(commandName, enableDebug, true, headers) {
				capabilityMapEntry := capabilityValue.(map[string]interface{})

				capabilitySetsMapList = append(capabilitySetsMapList, capabilityMapEntry)
			}
		} else {
			for _, capabilitySetConfig := range capabilitySetsConfig {
				capabilitySetConfigName := capabilitySetConfig.(string)

				for _, capabilityValue := range GetCapabilitySetsByName(commandName, enableDebug, true, capabilitySetConfigName, headers) {
					capabilityMapEntry := capabilityValue.(map[string]interface{})

					capabilitySetsMapList = append(capabilitySetsMapList, capabilityMapEntry)
				}
			}
		}

		var capabilitySetIds []string
		for _, mapEntry := range capabilitySetsMapList {
			capabilitySetId := mapEntry["id"].(string)

			capabilitySetIds = append(capabilitySetIds, capabilitySetId)
		}

		capabilitySetsBytes, err := json.Marshal(map[string]any{"roleId": roleId, "capabilitySetIds": capabilitySetIds})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, capabilitySetsBytes, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf(`Attached '%d' capability sets to '%s' role in '%s' realm`, len(capabilitySetIds), roleName, tenant))
	}
}
