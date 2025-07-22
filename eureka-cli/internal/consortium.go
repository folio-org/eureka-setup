package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// ######## Consortium ########

func GetConsortiumCentralTenant(consortium string, tenants map[string]any) string {
	for tenant, properties := range tenants {
		if properties == nil || !isValidConsortium(consortium, properties) {
			continue
		}
		if getSortableIsCentral(properties.(map[string]any)) == 0 {
			continue
		}

		return tenant
	}

	return ""
}

func getSortableIsCentral(mapEntry map[string]any) int {
	if GetBoolKey(mapEntry, TenantsCentralTenantEntryKey) {
		return 1
	}

	return 0
}

func isValidConsortium(consortium string, properties any) bool {
	return properties.(map[string]any)[TenantsConsortiumEntryKey] == consortium
}

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

// ######## Consortium Tenants ########

func GetSortedConsortiumTenants(consortium string, tenants map[string]any) ConsortiumTenants {
	var consortiumTenants ConsortiumTenants
	for tenant, properties := range tenants {
		if !isValidConsortium(consortium, properties) {
			continue
		}

		if properties == nil {
			consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: 0})
			continue
		}

		isCentral := getSortableIsCentral(properties.(map[string]any))
		consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: isCentral})
	}

	sort.Slice(consortiumTenants, func(i, j int) bool {
		return consortiumTenants[i].IsCentral > consortiumTenants[j].IsCentral
	})

	return consortiumTenants
}

func GetConsortiumUsers(consortium string, users map[string]any) map[string]any {
	consortiumUsers := make(map[string]any)
	for username, properties := range users {
		if !isValidConsortium(consortium, properties) {
			continue
		}

		consortiumUsers[username] = properties
	}

	return consortiumUsers
}

func GetAdminUsername(centralTenant string, consortiumUsers map[string]any) string {
	for username, properties := range consortiumUsers {
		tenant := properties.(map[string]any)[UsersTenantEntryKey]
		if tenant != nil && tenant.(string) == centralTenant {
			return username
		}
	}

	return ""
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

		var requestUrl = fmt.Sprintf("/consortia/%s/tenants", consortiumId)
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

	for _, value := range foundConsortiumTenantsMap["tenants"].([]any) {
		existingTenant := value.(map[string]any)["name"]
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

// ######## Central Ordering ########

func EnableCentralOrdering(commandName string, enableDebug bool, centralTenant string, accessToken string) {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"

	enableCentralOrdering := GetEnableCentralOrderingByKey(commandName, enableDebug, true, centralTenant, accessToken, centralOrderingLookupKey)
	if enableCentralOrdering {
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

func GetEnableCentralOrderingByKey(commandName string, enableDebug bool, panicOnError bool, centralTenant string, accessToken string, key string) bool {
	requestUrl := fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, fmt.Sprintf("/orders-storage/settings?query=key==%s", key))
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: centralTenant, TokenHeader: accessToken}

	foundSettingsMap := DoGetDecodeReturnMapStringAny(commandName, requestUrl, enableDebug, panicOnError, headers)
	if foundSettingsMap["settings"] == nil || len(foundSettingsMap["settings"].([]any)) == 0 {
		return false
	}

	settings := foundSettingsMap["settings"].([]any)[0]
	value := settings.(map[string]any)["value"].(string)

	enableCentralOrdering, err := strconv.ParseBool(value)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "strconv.ParseBool error")
		panic(err)
	}

	return enableCentralOrdering
}
