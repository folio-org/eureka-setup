package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	GatewayPort int = 8000

	JsonContentType           string = "application/json"
	FormUrlEncodedContentType string = "application/x-www-form-urlencoded"

	ContentTypeHeader   string = "Content-Type"
	AuthorizationHeader string = "Authorization"
	TenantHeader        string = "X-Okapi-Tenant"
	TokenHeader         string = "X-Okapi-Token"

	RemoveRoleUnsupported bool = true

	HealtcheckUri         string = "/admin/health"
	HealtcheckMaxAttempts int    = 50

	ModuleIdPattern string = "([a-z-_]+)([\\d-_.]+)([a-zA-Z0-9-_.]+)"
)

var (
	HealthcheckDefaultDuration time.Duration = 10 * time.Second

	ModuleIdRegexp *regexp.Regexp = regexp.MustCompile(ModuleIdPattern)
)

// ######## Auxiliary ########

func ExtractModuleNameAndVersion(commandName string, enableDebug bool, registryModulesMap map[string][]*RegistryModule, printOutput bool) {
	for registryName, registryModules := range registryModulesMap {
		if printOutput {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Extracting %s registry module names and versions", registryName))
		}

		for moduleIndex, module := range registryModules {
			if module.Id == "okapi" {
				continue
			}

			module.Name = TrimModuleName(ModuleIdRegexp.ReplaceAllString(module.Id, `$1`))
			module.Version = Stringp(ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`))

			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = fmt.Sprintf("%s-sc", module.Name)
			}

			registryModules[moduleIndex] = module
		}
	}
}

func PerformModuleHealthcheck(commandName string, enableDebug bool, waitMutex *sync.WaitGroup, moduleName string, port int) {
	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Waiting for module container %s on port %d to initialize", moduleName, port))

	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), port, HealtcheckUri)
	healthcheckAttempts := HealtcheckMaxAttempts
	for {
		time.Sleep(HealthcheckDefaultDuration)

		if checkContainerStatusCode(commandName, requestUrl, enableDebug) {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is healthy", moduleName))
			waitMutex.Done()
			break
		}

		healthcheckAttempts--

		if healthcheckAttempts == 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is unhealthy, out of attempts", moduleName))
			waitMutex.Done()
			LogErrorPanic(commandName, fmt.Sprintf("internal.PerformModuleHealthcheck - Module container %s did not initialize, cannot continue", moduleName))
		}

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Module container %s is unhealthy, %d/%d attempts left", moduleName, healthcheckAttempts, HealtcheckMaxAttempts))
	}
}

func checkContainerStatusCode(commandName string, requestUrl string, enableDebug bool) bool {
	response := DoGetReturnResponse(commandName, requestUrl, enableDebug, false, map[string]string{})
	if response == nil {
		return false
	}
	defer response.Body.Close()

	return response.StatusCode == 200
}

// ######## Application & Application Discovery ########

func GetApplications(commandName string, enableDebug bool, panicOnError bool) Applications {
	var applications Applications

	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/applications")

	response := DoGetReturnResponse(commandName, requestUrl, enableDebug, panicOnError, map[string]string{})
	if response == nil {
		return applications
	}
	defer response.Body.Close()

	err := json.NewDecoder(response.Body).Decode(&applications)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.NewDecoder error")
		panic(err)
	}

	return applications
}

func RemoveApplication(commandName string, enableDebug bool, panicOnError bool, applicationId string) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/applications/%s", applicationId))

	DoDelete(commandName, requestUrl, enableDebug, false, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Removed %s application", applicationId))
}

func CreateApplications(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []any
		frontendModuleDescriptors []any
		dependencies              map[string]any
		discoveryModules          []map[string]string
	)

	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)
	applicationPlatform := applicationMap["platform"].(string)
	applicationFetchDescriptors := applicationMap["fetch-descriptors"].(bool)

	applicationId := fmt.Sprintf("%s-%s", applicationName, applicationVersion)

	if applicationMap["dependencies"] != nil {
		dependencies = applicationMap["dependencies"].(map[string]any)
	}

	for registryName, registryModules := range dto.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Adding %s modules to %s application", registryName, applicationId))
		}

		for _, module := range registryModules {
			if strings.Contains(module.Name, ManagementModulePattern) {
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

			moduleDescriptorUrl := fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls[FolioRegistry], module.Id)

			if applicationFetchDescriptors {
				slog.Info(commandName, GetFuncName(), fmt.Sprintf("Fetching module descriptor for %s from %s", module.Id, moduleDescriptorUrl))
				dto.ModuleDescriptorsMap[module.Id] = DoGetDecodeReturnAny(commandName, moduleDescriptorUrl, enableDebug, true, map[string]string{})
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

			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Adding module to application, module: %s, version: %s", module.Name, *module.Version))
		}
	}

	applicationBytes, err := json.Marshal(map[string]any{
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

	DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/applications?check=true"), enableDebug, true, applicationBytes, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s application", applicationId))

	if len(discoveryModules) > 0 {
		applicationDiscoveryBytes, err := json.Marshal(map[string]any{"discovery": discoveryModules})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/modules/discovery"), enableDebug, true, applicationDiscoveryBytes, map[string]string{})
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %d entries of application module discovery", len(discoveryModules)))
}

func UpdateModuleDiscovery(commandName string, enableDebug bool, id string, sidecarUrl string, restore bool, portServer string) {
	id = strings.ReplaceAll(id, ":", "-")
	name := TrimModuleName(ModuleIdRegexp.ReplaceAllString(id, `$1`))
	if sidecarUrl == "" || restore {
		if strings.HasPrefix(name, "edge") {
			sidecarUrl = fmt.Sprintf("http://%s.eureka:%s", name, portServer)
		} else {
			sidecarUrl = fmt.Sprintf("http://%s-sc.eureka:%s", name, portServer)
		}
	}

	applicationDiscoveryBytes, err := json.Marshal(map[string]any{
		"id":       id,
		"name":     name,
		"version":  ModuleIdRegexp.ReplaceAllString(id, `$2$3`),
		"location": sidecarUrl,
	})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/modules/%s/discovery", id))

	DoPutReturnNoContent(commandName, requestUrl, enableDebug, applicationDiscoveryBytes, map[string]string{})

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Updated application module discovery for %s module with %s sidecar URL", name, sidecarUrl))
}

// ######## Tenants ########

func GetTenants(commandName string, enableDebug bool, panicOnError bool) []any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/tenants")

	foundTenantsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, map[string]string{})
	if foundTenantsMap["tenants"] == nil || len(foundTenantsMap["tenants"].([]any)) == 0 {
		return nil
	}

	return foundTenantsMap["tenants"].([]any)
}

func RemoveTenants(commandName string, enableDebug bool, panicOnError bool) {
	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant) {
			continue
		}

		requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Removed %s tenant (realm)", tenant))
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/tenants")
	tenants := ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey))

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{"name": tenant, "description": "Default"})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, true, tenantBytes, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s tenant (realm)", tenant))
	}
}

// ######## Tenant Entitlements ########

func RemoveTenantEntitlements(commandName string, enableDebug bool, panicOnError bool, purgeSchemas bool) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/entitlements?purge=%t&ignoreErrors=false", purgeSchemas))
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, panicOnError) {
		mapEntry := value.(map[string]any)

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

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Removed tenant entitlement for %s tenant (realm)", tenant))
	}
}

func CreateTenantEntitlement(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=loadReference=true,loadSample=true")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, false) {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant) {
			continue
		}

		applications := []string{fmt.Sprintf("%s-%s", applicationName, applicationVersion)}

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": mapEntry["id"].(string), "applications": applications})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, true, tenantEntitlementBytes, map[string]string{})

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created tenant entitlement for %s tenant (realm)", tenant))
	}
}

// ######## Users ########

func GetUser(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string, username string) any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/users?query=username==%s", username))
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	foundUsersMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundUsersMap["users"] == nil || len(foundUsersMap["users"].([]any)) == 0 {
		return nil
	}

	return foundUsersMap["users"].([]any)[0]
}

func GetUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) []any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/users?offset=0&limit=10000")
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	foundUsersMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundUsersMap["users"] == nil || len(foundUsersMap["users"].([]any)) == 0 {
		return nil
	}

	return foundUsersMap["users"].([]any)
}

func RemoveUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, value := range GetUsers(commandName, enableDebug, panicOnError, tenant, accessToken) {
		mapEntry := value.(map[string]any)

		username := mapEntry["username"].(string)
		usersMap := viper.GetStringMap(UsersKey)
		if usersMap[username] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Removed %s in %s tenant (realm)", username, tenant))
	}
}

func CreateUsers(commandName string, enableDebug bool, panicOnError bool, existingTenant string, accessToken string) {
	postUserRequestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/users-keycloak/users")
	postUserPasswordRequestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/authn/credentials")
	postUserRoleRequestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/roles/users")
	usersMap := viper.GetStringMap(UsersKey)

	for username, value := range usersMap {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		password := mapEntry["password"].(string)
		firstName := mapEntry["first-name"].(string)
		lastName := mapEntry["last-name"].(string)
		userRoles := mapEntry["roles"].([]any)

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

		createdUserMap := DoPostReturnMapStringAny(commandName, postUserRequestUrl, enableDebug, panicOnError, userBytes, okapiBasedHeaders)

		userId := createdUserMap["id"].(string)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s user with password %s in %s tenant (realm)", username, password, tenant))

		userPasswordBytes, err := json.Marshal(map[string]any{"userId": userId, "username": username, "password": password})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, postUserPasswordRequestUrl, enableDebug, panicOnError, userPasswordBytes, nonOkapiBasedHeaders)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Attached %s password to %s user in %s tenant (realm)", password, username, tenant))

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

		DoPostReturnNoContent(commandName, postUserRoleRequestUrl, enableDebug, panicOnError, userRoleBytes, okapiBasedHeaders)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Attached %d roles to %s user in %s tenant (realm)", len(roleIds), username, tenant))
	}
}

// ######## Roles ########

func GetRoles(commandName string, enableDebug bool, panicOnError bool, headers map[string]string) []any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/roles?offset=0&limit=10000")

	foundRolesMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundRolesMap["roles"] == nil || len(foundRolesMap["roles"].([]any)) == 0 {
		return nil
	}

	return foundRolesMap["roles"].([]any)
}

func GetRoleByName(commandName string, enableDebug bool, roleName string, headers map[string]string) map[string]any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/roles?query=name==%s", roleName))

	foundRolesMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, true, headers)
	if foundRolesMap["roles"] == nil {
		return nil
	}

	foundRoles := foundRolesMap["roles"].([]any)
	if len(foundRoles) != 1 {
		LogErrorPanic(commandName, fmt.Sprintf("internal.GetRoleByName error - Number of found roles by %s role name is not 1", roleName))
		return nil
	}

	return foundRoles[0].(map[string]any)
}

func RemoveRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}
	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(RolesKey)

	for _, value := range GetRoles(commandName, enableDebug, panicOnError, headers) {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if roles[roleName] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/roles/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Removed %s role in %s tenant (realm)", roleName, tenant))
	}
}

func CreateRoles(commandName string, enableDebug bool, panicOnError bool, existingTenant string, accessToken string) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/roles")
	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(RolesKey)

	for role, value := range roles {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

		roleBytes, err := json.Marshal(map[string]string{"name": caser.String(role), "description": "Default"})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, panicOnError, roleBytes, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s role in %s tenant (realm)", role, tenant))
	}
}

// ######## Capabilities ########

func GetCapabilitySets(commandName string, enableDebug bool, panicOnError bool, headers map[string]string) []any {
	var foundCapabilitySets []any

	applications := GetApplications(commandName, enableDebug, panicOnError)

	for _, value := range applications.ApplicationDescriptors {
		applicationId := value["id"].(string)

		requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationId))

		foundCapabilitySetsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
		if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
			return nil
		}

		foundCapabilitySets = append(foundCapabilitySets, foundCapabilitySetsMap["capabilitySets"].([]any)...)
	}

	return foundCapabilitySets
}

func GetCapabilitySetsByName(commandName string, enableDebug bool, panicOnError bool, headers map[string]string, capabilitySetName string) []any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilitySetName))

	foundCapabilitySetsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
		return nil
	}

	return foundCapabilitySetsMap["capabilitySets"].([]any)
}

func DetachCapabilitySetsFromRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}
	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(RolesKey)

	for _, value := range GetRoles(commandName, enableDebug, panicOnError, headers) {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, false, headers)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Detached capability sets from %s role in %s tenant (realm)", roleName, tenant))
	}
}

func AttachCapabilitySetsToRoles(commandName string, enableDebug bool, tenant string, accessToken string) {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/roles/capability-sets")
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}
	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(RolesKey)

	for _, roleValue := range GetRoles(commandName, enableDebug, true, headers) {
		mapEntry := roleValue.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		rolesMapConfig := rolesMap[roleName].(map[string]any)
		if tenant != rolesMapConfig[RolesTenantEntryKey].(string) {
			continue
		}

		capabilitySetIds := populateCapabilitySets(commandName, enableDebug, headers, rolesMapConfig[RolesCapabilitySetsEntryKey].([]any))
		if len(capabilitySetIds) == 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("No capability sets were attached to %s role in %s tenant (realm)", roleName, tenant))
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySetIds); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySetIds))
			batchCapabilitySetIds := capabilitySetIds[lowerBound:upperBound]

			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Attaching %d-%d (total: %d) capability sets to %s role in %s tenant (realm)", lowerBound, upperBound, len(capabilitySetIds), roleName, tenant))

			capabilitySetsBytes, err := json.Marshal(map[string]any{"roleId": mapEntry["id"].(string), "capabilitySetIds": batchCapabilitySetIds})
			if err != nil {
				slog.Error(commandName, GetFuncName(), "json.Marshal error")
				panic(err)
			}

			DoRetryablePostReturnNoContent(commandName, requestUrl, enableDebug, true, capabilitySetsBytes, headers)
		}

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Attached %d capability sets to %s role in %s tenant (realm)", len(capabilitySetIds), roleName, tenant))
	}
}

func populateCapabilitySets(commandName string, enableDebug bool, headers map[string]string, capabilitySetNames []any) []string {
	var capabilitySets []string = []string{}
	if len(capabilitySetNames) > 1 && !slices.Contains(capabilitySetNames, "all") {
		for _, capabilitySetName := range capabilitySetNames {
			for _, value := range GetCapabilitySetsByName(commandName, enableDebug, true, headers, capabilitySetName.(string)) {
				capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
			}
		}
	} else {
		for _, value := range GetCapabilitySets(commandName, enableDebug, true, headers) {
			capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
		}
	}

	return capabilitySets
}

// ######## Consortium ########

func CreateConsortium(commandName string, enableDebug bool, centralTenant string, accessToken string, consortiumName string) string {
	consortium := GetConsortiumByName(commandName, enableDebug, true, centralTenant, accessToken, consortiumName)
	if consortium != nil {
		consortiumId := consortium.(map[string]any)["id"].(string)

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Consortium %s is already created", consortiumName))

		return consortiumId
	}

	consortiumId := uuid.New()

	bytes, err := json.Marshal(map[string]any{"id": consortiumId, "name": consortiumName})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/consortia"), enableDebug, true, bytes, headers)

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s consortium", consortiumName))

	return consortiumId.String()
}

func GetConsortiumByName(commandName string, enableDebug bool, panicOnError bool, centralTenant string, accessToken string, consortiumName string) any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/consortia?query=name==%s", consortiumName))
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	foundConsortiumsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundConsortiumsMap["consortia"] == nil || len(foundConsortiumsMap["consortia"].([]any)) == 0 {
		return nil
	}

	return foundConsortiumsMap["consortia"].([]any)[0]
}

func CreateConsortiumTenants(commandName string, enableDebug bool, centralTenant string, accessToken string, consortiumId string, consortiumTenants ConsortiumTenants, adminUsername string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	for _, consortiumTenant := range consortiumTenants {
		bytes, err := json.Marshal(map[string]any{
			"id":        consortiumTenant.Tenant,
			"code":      consortiumTenant.Tenant[0:3],
			"name":      consortiumTenant.Tenant,
			"isCentral": consortiumTenant.IsCentral,
		})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		existingTenant := GetConsortiumTenantByIdAndName(commandName, enableDebug, true, centralTenant, accessToken, consortiumId, consortiumTenant.Tenant)
		if existingTenant != nil {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Consortium tenant %s is already created", consortiumTenant.Tenant))
			continue
		}

		var requestUrl string = fmt.Sprintf("/consortia/%s/tenants", consortiumId)
		if consortiumTenant.IsCentral == 0 {
			user := GetUser(commandName, enableDebug, true, centralTenant, accessToken, adminUsername)

			requestUrl = fmt.Sprintf("/consortia/%s/tenants?adminUserId=%s", consortiumId, user.(map[string]any)["id"].(string))
		}

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Trying to create %s consortium tenant for %s consortium", consortiumTenant.Tenant, consortiumId))

		DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, requestUrl), enableDebug, true, bytes, headers)

		CheckConsortiumTenantStatus(commandName, enableDebug, true, centralTenant, consortiumId, consortiumTenant.Tenant, headers)
	}
}

func GetConsortiumTenantByIdAndName(commandName string, enableDebug bool, panicOnError bool, centralTenant string, accessToken string, consortiumId string, tenant string) any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/consortia/%s/tenants", consortiumId))
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	foundConsortiumTenantsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundConsortiumTenantsMap["tenants"] == nil || len(foundConsortiumTenantsMap["tenants"].([]any)) == 0 {
		return nil
	}

	existingTenants := foundConsortiumTenantsMap["tenants"].([]any)

	for _, value := range existingTenants {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"]
		if existingTenant != nil && existingTenant.(string) == tenant {
			return existingTenant
		}
	}

	return nil
}

func CheckConsortiumTenantStatus(commandName string, enableDebug bool, panicOnError bool, centralTenant string, consortiumId string, tenant string, headers map[string]string) {
	requestUrl := fmt.Sprintf("/consortia/%s/tenants/%s", consortiumId, tenant)

	foundConsortiumTenantMap := DoGetDecodeReturnMapStringAny(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, requestUrl), enableDebug, true, headers)
	if foundConsortiumTenantMap == nil {
		return
	}

	const (
		IN_PROGRESS           string = "IN_PROGRESS"
		FAILED                string = "FAILED"
		COMPLETED             string = "COMPLETED"
		COMPLETED_WITH_ERRORS string = "COMPLETED_WITH_ERRORS"

		WaitConsortiumTenant time.Duration = 10 * time.Second
	)

	switch foundConsortiumTenantMap["setupStatus"] {
	case IN_PROGRESS:
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Waiting for %s consortium tenant creation", tenant))
		time.Sleep(WaitConsortiumTenant)
		CheckConsortiumTenantStatus(commandName, enableDebug, true, centralTenant, consortiumId, tenant, headers)
		return
	case FAILED:
	case COMPLETED_WITH_ERRORS:
		LogErrorPanic(commandName, fmt.Sprintf("CheckConsortiumTenantStatus() error - %s consortium tenant not is created", tenant))
		return
	case COMPLETED:
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Created %s consortium tenant (%t) for %s consortium", tenant, foundConsortiumTenantMap["isCentral"], consortiumId))
		return
	}
}

func EnableCentralOrdering(commandName string, enableDebug bool, centralTenant string, accessToken string) {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"

	enableCentralOrdering := GetEnableCentralOrderingByKey(commandName, enableDebug, true, centralTenant, accessToken, centralOrderingLookupKey)
	if enableCentralOrdering != nil && enableCentralOrdering.(map[string]any)["value"].(string) == "true" {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Central ordering for %s tenant is already enabled", centralTenant))
		return
	}

	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	bytes, err := json.Marshal(map[string]any{"key": centralOrderingLookupKey, "value": "true"})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/orders-storage/settings"), enableDebug, true, bytes, headers)

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Enabled central ordering for %s tenant", centralTenant))
}

func GetEnableCentralOrderingByKey(commandName string, enableDebug bool, panicOnError bool, centralTenant string, accessToken string, key string) any {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/orders-storage/settings?query=key==%s", key))
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	foundConsortiumsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundConsortiumsMap["settings"] == nil || len(foundConsortiumsMap["settings"].([]any)) == 0 {
		return nil
	}

	return foundConsortiumsMap["settings"].([]any)[0]
}
