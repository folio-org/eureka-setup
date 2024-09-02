package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const (
	GatewayPort = 8000

	ApplicationsPort       = 9901
	TenantsPort            = 9902
	TenantEntitlementsPort = 9903

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
	resp := DoGetReturnResponse(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), enableDebug, panicOnError, map[string]string{})
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

		DoDelete(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, fmt.Sprintf("/applications/%s", id)), enableDebug, map[string]string{})
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
				dto.ModuleDescriptorsMap[module.Id] = DoGetDecodeReturnInterface(commandName, moduleDescriptorUrl, enableDebug, true, map[string]string{})
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

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications?check=true"), enableDebug, applicationBytes, map[string]string{})

	applicationDiscoveryBytes, err := json.Marshal(map[string]interface{}{"discovery": discoveryModules})
	if err != nil {
		slog.Error(commandName, "json.Marshal error", "")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), enableDebug, applicationDiscoveryBytes, map[string]string{})
}

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

		if !slices.Contains(viper.GetStringSlice(TenantsKey), tenant) {
			continue
		}

		requestUrl := fmt.Sprintf("/tenants/%s?purge=true", mapEntry["id"].(string))

		DoDelete(commandName, fmt.Sprintf(DockerInternalUrl, TenantsPort, requestUrl), enableDebug, map[string]string{})

		slog.Info(commandName, fmt.Sprintf("Removed tenant %s", tenant), "")
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

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantBytes, map[string]string{})

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
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(TenantsKey), tenant) {
			continue
		}

		tenantId := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		DoDeleteWithBody(commandName, requestUrl, enableDebug, tenantEntitlementBytes, true, map[string]string{})

		slog.Info(commandName, fmt.Sprintf("Removed tenant entitlement %s tenant", tenant), "")

	}
}

func CreateTenantEntitlement(commandName string, enableDebug bool) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, TenantEntitlementsPort, "/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=loadReference=false,loadSample=false")
	applicationMap := viper.GetStringMap(ApplicationKey)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range GetTenants(commandName, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(TenantsKey), tenant) {
			continue
		}

		tenantId := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, tenantEntitlementBytes, map[string]string{})

		slog.Info(commandName, fmt.Sprintf("Created tenant entitlement for %s tenant (%s)", tenant, tenantId), "")

	}
}

func GetUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) []interface{} {
	var foundUsers []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/users")

	headers := map[string]string{
		ContentTypeHeader: JsonContentType,
		TenantHeader:      tenant,
		TokenHeader:       accessToken,
	}

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundTenantsMap["users"] == nil || len(foundTenantsMap["users"].([]interface{})) == 0 {
		return nil
	}

	foundUsers = foundTenantsMap["users"].([]interface{})

	return foundUsers
}

func RemoveUsers(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	for _, value := range GetUsers(commandName, enableDebug, panicOnError, tenant, accessToken) {
		mapEntry := value.(map[string]interface{})
		username := mapEntry["username"].(string)

		usersMap := viper.GetStringMap(UsersKey)
		if usersMap[username] == nil {
			continue
		}

		headers := map[string]string{
			ContentTypeHeader: JsonContentType,
			TenantHeader:      tenant,
			TokenHeader:       accessToken,
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, headers)

		slog.Info(commandName, fmt.Sprintf("Removed user %s", username), "")
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
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		okapiBasedHeaders := map[string]string{
			ContentTypeHeader: JsonContentType,
			TenantHeader:      tenant,
			TokenHeader:       accessToken,
		}

		nonOkapiBasedHeaders := map[string]string{
			ContentTypeHeader:   JsonContentType,
			TenantHeader:        tenant,
			AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
		}

		createdUserMap := DoPostReturnMapStringInteface(commandName, postUserRequestUrl, enableDebug, userBytes, okapiBasedHeaders)

		userId := createdUserMap["id"].(string)

		slog.Info(commandName, fmt.Sprintf("Created user %s (%s) with password %s in %s realm", username, userId, password, tenant), "")

		userPasswordBytes, err := json.Marshal(map[string]any{"userId": userId, "username": username, "password": password})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		DoPostReturnNoContent(commandName, postUserPasswordRequestUrl, enableDebug, userPasswordBytes, nonOkapiBasedHeaders)

		slog.Info(commandName, fmt.Sprintf("Attached password %s to user %s (%s) in %s realm", password, username, userId, tenant), "")

		var roleIds []string
		for _, userRole := range userRoles {
			role := GetRoleByName(commandName, enableDebug, userRole.(string), okapiBasedHeaders)

			roleId := role["id"].(string)
			roleName := role["name"].(string)

			if roleId == "" {
				slog.Warn(commandName, fmt.Sprintf("internal.GetRoleIdByName warn - Did not find role %s by name", roleName), "")
				continue
			}

			roleIds = append(roleIds, roleId)
		}

		userRoleBytes, err := json.Marshal(map[string]any{"userId": userId, "roleIds": roleIds})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		DoPostReturnNoContent(commandName, postUserRoleRequestUrl, enableDebug, userRoleBytes, okapiBasedHeaders)

		slog.Info(commandName, fmt.Sprintf("Attached %d roles to user %s (%s) in %s realm", len(roleIds), username, userId, tenant), "")
	}
}

func GetRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) []interface{} {
	var foundRoles []interface{}

	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles")

	headers := map[string]string{
		ContentTypeHeader: JsonContentType,
		TenantHeader:      tenant,
		TokenHeader:       accessToken,
	}

	foundTenantsMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundTenantsMap["roles"] == nil || len(foundTenantsMap["roles"].([]interface{})) == 0 {
		return nil
	}

	foundRoles = foundTenantsMap["roles"].([]interface{})

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
		LogErrorPanic(commandName, fmt.Sprintf("internal.DoGetDecodeReturnMapStringInteface - Found more than 1 role by %s role name", roleName))
	}

	return foundRoles[0].(map[string]interface{})
}

func RemoveRoles(commandName string, enableDebug bool, panicOnError bool, tenant string, accessToken string) {
	for _, value := range GetRoles(commandName, enableDebug, panicOnError, tenant, accessToken) {
		mapEntry := value.(map[string]interface{})
		roleName := mapEntry["name"].(string)

		rolesMap := viper.GetStringMap(RolesKey)
		if rolesMap[roleName] == nil {
			continue
		}

		headers := map[string]string{
			ContentTypeHeader: JsonContentType,
			TenantHeader:      tenant,
			TokenHeader:       accessToken,
		}

		requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, fmt.Sprintf("/roles-keycloak/roles/%s", mapEntry["id"].(string)))

		DoDelete(commandName, requestUrl, enableDebug, headers)

		slog.Info(commandName, fmt.Sprintf("Removed role %s", roleName), "")
	}
}

func CreateRoles(commandName string, enableDebug bool, accessToken string) {
	requestUrl := fmt.Sprintf(DockerInternalUrl, GatewayPort, "/roles")
	rolesMap := viper.GetStringMap(RolesKey)

	for role, value := range rolesMap {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["tenant"].(string)

		roleBytes, err := json.Marshal(map[string]string{"name": strings.Title(role), "description": "Default"})
		if err != nil {
			slog.Error(commandName, "json.Marshal error", "")
			panic(err)
		}

		headers := map[string]string{
			ContentTypeHeader: JsonContentType,
			TenantHeader:      tenant,
			TokenHeader:       accessToken,
		}

		DoPostReturnNoContent(commandName, requestUrl, enableDebug, roleBytes, headers)

		slog.Info(commandName, fmt.Sprintf("Created %s role", role), "")
	}
}

func PerformModuleHealthcheck(commandName string, enableDebug bool, waitMutex *sync.WaitGroup, moduleName string, port int) {
	slog.Info(commandName, fmt.Sprintf("Waiting for module container %s on port %d to initialize", moduleName, port), "")
	requestUrl := fmt.Sprintf(DockerInternalUrl, port, HealtcheckUri)
	healthcheckAttempts := HealtcheckMaxAttempts
	for {
		time.Sleep(HealthcheckDefaultDuration)

		isHealthyVertx := false
		actuatorHealthStr := DoGetDecodeReturnString(commandName, requestUrl, enableDebug, false, map[string]string{})
		if strings.Contains(actuatorHealthStr, "OK") {
			isHealthyVertx = !isHealthyVertx
		}

		isHealthySpringBoot := false
		actuatorHealthMap := DoGetDecodeReturnMapStringInteface(commandName, requestUrl, enableDebug, false, map[string]string{})
		if actuatorHealthMap != nil && strings.Contains(actuatorHealthMap["status"].(string), "UP") {
			isHealthySpringBoot = !isHealthySpringBoot
		}

		if isHealthyVertx || isHealthySpringBoot {
			slog.Info(commandName, fmt.Sprintf("Module container %s is healthy", moduleName), "")
			waitMutex.Done()
			break
		}

		healthcheckAttempts--
		if healthcheckAttempts > 0 {
			slog.Info(commandName, fmt.Sprintf("Module container %s is unhealthy, %d/%d attempts left", moduleName, healthcheckAttempts, HealtcheckMaxAttempts), "")
		} else {
			slog.Info(commandName, fmt.Sprintf("Module container %s is unhealthy, out of attempts", moduleName), "")
			waitMutex.Done()
			break
		}
	}
}
