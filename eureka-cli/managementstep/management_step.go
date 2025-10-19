package managementstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/tenantstep"
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ManagementStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	TenantStep *tenantstep.TenantStep
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, tenantStep *tenantstep.TenantStep) *ManagementStep {
	return &ManagementStep{
		Action:     action,
		HTTPClient: httpClient,
		TenantStep: tenantStep,
	}
}

// ######## Application & Application Discovery ########

func (ms *ManagementStep) GetApplications(panicOnError bool) models.Applications {
	var applications models.Applications

	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/applications")

	response := ms.HTTPClient.DoGetReturnResponse(requestURL, panicOnError, map[string]string{})
	if response == nil {
		return applications
	}
	defer func() {
		_ = response.Body.Close()
	}()

	err := json.NewDecoder(response.Body).Decode(&applications)
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	return applications
}

func (ms *ManagementStep) CreateApplications(extract *models.RegistryModuleExtract) {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []any
		frontendModuleDescriptors []any
		dependencies              map[string]any
		discoveryModules          []map[string]string
	)

	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)
	applicationPlatform := applicationMap["platform"].(string)
	applicationFetchDescriptors := applicationMap["fetch-descriptors"].(bool)

	applicationId := fmt.Sprintf("%s-%s", applicationName, applicationVersion)

	if applicationMap["dependencies"] != nil {
		dependencies = applicationMap["dependencies"].(map[string]any)
	}

	for registryName, registryModules := range extract.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Adding %s modules to %s application", registryName, applicationId))
		}

		for _, module := range registryModules {
			if strings.Contains(module.Name, constant.ManagementModulePattern) {
				continue
			}

			backendModule, okBackend := extract.BackendModulesMap[module.Name]
			frontendModule, okFrontend := extract.FrontendModulesMap[module.Name]
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

			moduleDescriptorUrl := fmt.Sprintf("%s/_/proxy/modules/%s", extract.RegistryURLs[constant.FolioRegistry], module.Id)

			isLocalModule := okBackend && backendModule.LocalDescriptorPath != ""

			if applicationFetchDescriptors || isLocalModule {
				if isLocalModule {
					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Fetching local module descriptor for %s from file path", module.Id))

					var descriptor map[string]any
					helpers.ReadJsonFromFile(ms.Action, backendModule.LocalDescriptorPath, &descriptor)
					extract.ModuleDescriptorsMap[module.Id] = descriptor

					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Successfully loaded descriptor for %s from local file", module.Id))
				} else {
					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Fetching module descriptor for %s from %s", module.Id, moduleDescriptorUrl))
					extract.ModuleDescriptorsMap[module.Id] = ms.HTTPClient.DoGetDecodeReturnAny(moduleDescriptorUrl, true, map[string]string{})
				}
			}

			if okBackend {
				serverPort := strconv.Itoa(backendModule.ModuleServerPort)
				backendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}

				if applicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptorsMap[module.Id])
				} else {
					backendModule["url"] = moduleDescriptorUrl
				}

				backendModules = append(backendModules, backendModule)

				sidecarUrl := fmt.Sprintf("http://%s.eureka:%s", module.SidecarName, serverPort)

				discoveryModules = append(discoveryModules, map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version, "location": sidecarUrl})
			} else if okFrontend {
				frontendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}
				if applicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, extract.ModuleDescriptorsMap[module.Id])
				} else {
					frontendModule["url"] = moduleDescriptorUrl
				}

				frontendModules = append(frontendModules, frontendModule)
			}

			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Adding module to application, module: %s, version: %s", module.Name, *module.Version))
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
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	ms.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/applications?check=true"), true, applicationBytes, map[string]string{})

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %s application", applicationId))

	if len(discoveryModules) > 0 {
		applicationDiscoveryBytes, err := json.Marshal(map[string]any{"discovery": discoveryModules})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/modules/discovery"), true, applicationDiscoveryBytes, map[string]string{})
	}

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %d entries of application module discovery", len(discoveryModules)))
}

func (ms *ManagementStep) RemoveApplication(panicOnError bool, applicationId string) {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/applications/%s", applicationId))

	ms.HTTPClient.DoDelete(requestURL, false, map[string]string{})

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed %s application", applicationId))
}

func (ms *ManagementStep) UpdateModuleDiscovery(id string, sidecarUrl string, restore bool, portServer string) {
	id = strings.ReplaceAll(id, ":", "-")
	name := helpers.GetModuleName(id)
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
		"version":  helpers.GetModuleVersion(id),
		"location": sidecarUrl,
	})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/modules/%s/discovery", id))

	ms.HTTPClient.DoPutReturnNoContent(requestURL, applicationDiscoveryBytes, map[string]string{})

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Updated application module discovery for %s module with %s sidecar URL", name, sidecarUrl))
}

// ######## Tenants ########

func (ms *ManagementStep) GetTenants(panicOnError bool, consortiumName string, tenantType tenanttype.TenantType) []any {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/tenants")
	if tenantType != tenanttype.All {
		requestURL += fmt.Sprintf("?query=description==%s-%s", consortiumName, tenantType)
	}

	foundTenantsMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, map[string]string{})
	if foundTenantsMap["tenants"] == nil || len(foundTenantsMap["tenants"].([]any)) == 0 {
		return nil
	}

	return foundTenantsMap["tenants"].([]any)
}

func (ms *ManagementStep) CreateTenants() {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/tenants")
	tenants := viper.GetStringMap(field.Tenants)

	for tenant, properties := range tenants {
		mapEntry := properties.(map[string]any)

		description := fmt.Sprintf("%s-%s", constant.NoneConsortium, tenanttype.Default)

		consortiumName := helpers.GetAnyOrDefault(mapEntry, field.TenantsConsortiumEntry, nil)
		if consortiumName != nil {
			tenantType := tenanttype.Member
			if helpers.GetBoolKey(mapEntry, field.TenantsCentralTenantEntry) {
				tenantType = tenanttype.Central
			}

			description = fmt.Sprintf("%s-%s", consortiumName, tenantType)
		}

		tenantBytes, err := json.Marshal(map[string]string{"name": tenant, "description": description})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(requestURL, true, tenantBytes, map[string]string{})

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %s tenant (realm) %s description", tenant, description))
	}
}

func (ms *ManagementStep) RemoveTenants(panicOnError bool, consortiumName string, tenantType tenanttype.TenantType) {
	for _, value := range ms.GetTenants(panicOnError, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(viper.GetStringMap(field.Tenants)), tenant) {
			continue
		}

		requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", mapEntry["id"].(string)))

		ms.HTTPClient.DoDelete(requestURL, false, map[string]string{})

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed %s tenant (realm)", tenant))
	}
}

// ######## Tenant Entitlements ########

func (ms *ManagementStep) CreateTenantEntitlement(consortiumName string, tenantType tenanttype.TenantType) {
	tenants := viper.GetStringMap(field.Tenants)

	tenantParameters := ms.TenantStep.GetTenantParameters(consortiumName, tenants)

	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=%s", tenantParameters))
	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range ms.GetTenants(false, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(tenants), tenant) {
			continue
		}

		applications := []string{fmt.Sprintf("%s-%s", applicationName, applicationVersion)}

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": mapEntry["id"].(string), "applications": applications})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(requestURL, true, tenantEntitlementBytes, map[string]string{})

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created tenant entitlement for %s tenant (realm)", tenant))
	}
}

func (ms *ManagementStep) RemoveTenantEntitlements(panicOnError bool, purgeSchemas bool, consortiumName string, tenantType tenanttype.TenantType) {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/entitlements?purge=%t&ignoreErrors=false", purgeSchemas))
	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	for _, value := range ms.GetTenants(panicOnError, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(viper.GetStringMap(field.Tenants)), tenant) {
			continue
		}

		tenantId := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		tenantEntitlementBytes, err := json.Marshal(map[string]any{"tenantId": tenantId, "applications": applications})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoDeleteWithBody(requestURL, tenantEntitlementBytes, true, map[string]string{})

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed tenant entitlement for %s tenant (realm)", tenant))
	}
}

// ######## Users ########

func (ms *ManagementStep) GetUsers(panicOnError bool, tenant string, accessToken string) []any {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/users?offset=0&limit=10000")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	foundUsersMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if foundUsersMap["users"] == nil || len(foundUsersMap["users"].([]any)) == 0 {
		return nil
	}

	return foundUsersMap["users"].([]any)
}

func (ms *ManagementStep) CreateUsers(panicOnError bool, existingTenant string, accessToken string) {
	postUserRequestUrl := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/users-keycloak/users")
	postUserPasswordRequestUrl := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/authn/credentials")
	postUserRoleRequestUrl := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/roles/users")
	usersMap := viper.GetStringMap(field.Users)

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
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		okapiBasedHeaders := map[string]string{
			constant.ContentTypeHeader: constant.JsonContentType,
			constant.TenantHeader:      tenant,
			constant.TokenHeader:       accessToken,
		}

		nonOkapiBasedHeaders := map[string]string{
			constant.ContentTypeHeader:   constant.JsonContentType,
			constant.TenantHeader:        tenant,
			constant.AuthorizationHeader: fmt.Sprintf("Bearer %s", accessToken),
		}

		createdUserMap := ms.HTTPClient.DoPostReturnMapStringAny(postUserRequestUrl, panicOnError, userBytes, okapiBasedHeaders)

		userId := createdUserMap["id"].(string)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %s user with password %s in %s tenant (realm)", username, password, tenant))

		userPasswordBytes, err := json.Marshal(map[string]any{"userId": userId, "username": username, "password": password})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(postUserPasswordRequestUrl, panicOnError, userPasswordBytes, nonOkapiBasedHeaders)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Attached %s password to %s user in %s tenant (realm)", password, username, tenant))

		var roleIds []string
		for _, userRole := range userRoles {
			role := ms.GetRoleByName(userRole.(string), okapiBasedHeaders)
			roleId := role["id"].(string)
			roleName := role["name"].(string)

			if roleId == "" {
				slog.Warn(ms.Action.Name, "text", fmt.Sprintf("did not find role %s by name", roleName))
				continue
			}

			roleIds = append(roleIds, roleId)
		}

		userRoleBytes, err := json.Marshal(map[string]any{"userId": userId, "roleIds": roleIds})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(postUserRoleRequestUrl, panicOnError, userRoleBytes, okapiBasedHeaders)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Attached %d roles to %s user in %s tenant (realm)", len(roleIds), username, tenant))
	}
}

func (ms *ManagementStep) RemoveUsers(panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	for _, value := range ms.GetUsers(panicOnError, tenant, accessToken) {
		mapEntry := value.(map[string]any)

		username := mapEntry["username"].(string)
		usersMap := viper.GetStringMap(field.Users)
		if usersMap[username] == nil {
			continue
		}

		requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/users-keycloak/users/%s", mapEntry["id"].(string)))

		ms.HTTPClient.DoDelete(requestURL, false, headers)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed %s in %s tenant (realm)", username, tenant))
	}
}

// ######## Roles ########

func (ms *ManagementStep) GetRoles(panicOnError bool, headers map[string]string) []any {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/roles?offset=0&limit=10000")

	foundRolesMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if foundRolesMap["roles"] == nil || len(foundRolesMap["roles"].([]any)) == 0 {
		return nil
	}

	return foundRolesMap["roles"].([]any)
}

func (ms *ManagementStep) GetRoleByName(roleName string, headers map[string]string) map[string]any {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/roles?query=name==%s", roleName))

	foundRolesMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, true, headers)
	if foundRolesMap["roles"] == nil {
		return nil
	}

	foundRoles := foundRolesMap["roles"].([]any)
	if len(foundRoles) != 1 {
		helpers.LogErrorPanic(ms.Action, fmt.Errorf("number of found roles by %s role name is not 1", roleName))
		return nil
	}

	return foundRoles[0].(map[string]any)
}

func (ms *ManagementStep) CreateRoles(panicOnError bool, existingTenant string, accessToken string) {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/roles")
	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(field.Roles)

	for role, value := range roles {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["tenant"].(string)
		if existingTenant != tenant {
			continue
		}

		headers := map[string]string{
			constant.ContentTypeHeader: constant.JsonContentType,
			constant.TenantHeader:      tenant,
			constant.TokenHeader:       accessToken,
		}

		roleBytes, err := json.Marshal(map[string]string{"name": caser.String(role), "description": "Default"})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		ms.HTTPClient.DoPostReturnNoContent(requestURL, panicOnError, roleBytes, headers)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %s role in %s tenant (realm)", role, tenant))
	}
}

func (ms *ManagementStep) RemoveRoles(panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	caser := cases.Lower(language.English)
	roles := viper.GetStringMap(field.Roles)

	for _, value := range ms.GetRoles(panicOnError, headers) {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if roles[roleName] == nil {
			continue
		}

		requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/roles/%s", mapEntry["id"].(string)))

		ms.HTTPClient.DoDelete(requestURL, false, headers)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed %s role in %s tenant (realm)", roleName, tenant))
	}
}

// ######## Capabilities ########

func (ms *ManagementStep) GetCapabilitySets(panicOnError bool, headers map[string]string) []any {
	var foundCapabilitySets []any

	applications := ms.GetApplications(panicOnError)

	for _, value := range applications.ApplicationDescriptors {
		applicationId := value["id"].(string)

		requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationId))

		foundCapabilitySetsMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
		if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
			return nil
		}

		foundCapabilitySets = append(foundCapabilitySets, foundCapabilitySetsMap["capabilitySets"].([]any)...)
	}

	return foundCapabilitySets
}

func (ms *ManagementStep) GetCapabilitySetsByName(panicOnError bool, headers map[string]string, capabilitySetName string) []any {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilitySetName))

	foundCapabilitySetsMap := ms.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
		return nil
	}

	return foundCapabilitySetsMap["capabilitySets"].([]any)
}

func (ms *ManagementStep) AttachCapabilitySetsToRoles(tenant string, accessToken string) {
	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/roles/capability-sets")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(field.Roles)

	foundRoles := ms.GetRoles(true, headers)
	if len(foundRoles) == 0 {
		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Cannot attach capability sets, found no roles in %s tenant (realm)", tenant))
		return
	}

	for _, roleValue := range foundRoles {
		mapEntry := roleValue.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		rolesMapConfig := rolesMap[roleName].(map[string]any)
		if tenant != rolesMapConfig[field.RolesTenantEntry].(string) {
			continue
		}

		capabilitySetIds := ms.populateCapabilitySets(headers, rolesMapConfig[field.RolesCapabilitySetsEntry].([]any))
		if len(capabilitySetIds) == 0 {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("No capability sets were attached to %s role in %s tenant (realm)", roleName, tenant))
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySetIds); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySetIds))
			batchCapabilitySetIds := capabilitySetIds[lowerBound:upperBound]

			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Attaching %d-%d (total: %d) capability sets to %s role in %s tenant (realm)", lowerBound, upperBound, len(capabilitySetIds), roleName, tenant))

			capabilitySetsBytes, err := json.Marshal(map[string]any{"roleId": mapEntry["id"].(string), "capabilitySetIds": batchCapabilitySetIds})
			if err != nil {
				slog.Error(ms.Action.Name, "error", err)
				panic(err)
			}

			ms.HTTPClient.DoRetryPostReturnNoContent(requestURL, true, capabilitySetsBytes, headers)
		}

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Attached %d capability sets to %s role in %s tenant (realm)", len(capabilitySetIds), roleName, tenant))
	}
}

func (ms *ManagementStep) populateCapabilitySets(headers map[string]string, capabilitySetNames []any) []string {
	var capabilitySets = []string{}
	if len(capabilitySetNames) == 0 {
		return capabilitySets
	}

	if len(capabilitySetNames) == 1 && !slices.Contains(capabilitySetNames, "all") {
		for _, capabilitySetName := range capabilitySetNames {
			for _, value := range ms.GetCapabilitySetsByName(true, headers, capabilitySetName.(string)) {
				capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
			}

		}

		return capabilitySets
	}

	for _, value := range ms.GetCapabilitySets(true, headers) {
		capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
	}

	return capabilitySets
}

func (ms *ManagementStep) DetachCapabilitySetsFromRoles(panicOnError bool, tenant string, accessToken string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(field.Roles)

	foundRoles := ms.GetRoles(panicOnError, headers)
	if len(foundRoles) == 0 {
		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Cannot detach capability sets, found no roles in %s tenant (realm)", tenant))
		return
	}

	for _, value := range foundRoles {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		ms.HTTPClient.DoDelete(requestURL, false, headers)

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Detached capability sets from %s role in %s tenant (realm)", roleName, tenant))
	}
}
