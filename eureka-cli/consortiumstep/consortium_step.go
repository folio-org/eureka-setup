package consortiumstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/userstep"
	"github.com/google/uuid"
)

type ConsortiumStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	UserStep   *userstep.UserStep
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, userStep *userstep.UserStep) *ConsortiumStep {
	return &ConsortiumStep{
		Action:     action,
		HTTPClient: httpClient,
		UserStep:   userStep,
	}
}

// ######## Consortium ########

func (cs *ConsortiumStep) GetConsortiumCentralTenant(consortiumName string, tenants map[string]any) string {
	for tenant, properties := range tenants {
		if properties == nil || !cs.isValidConsortium(consortiumName, properties) || cs.getSortableIsCentral(properties.(map[string]any)) == 0 {
			continue
		}

		return tenant
	}

	return ""
}

func (cs *ConsortiumStep) getSortableIsCentral(mapEntry map[string]any) int {
	if helpers.GetBoolKey(mapEntry, field.TenantsCentralTenantEntry) {
		return 1
	}

	return 0
}

func (cs *ConsortiumStep) isValidConsortium(consortiumName string, properties any) bool {
	return properties.(map[string]any)[field.TenantsConsortiumEntry] == consortiumName
}

func (cs *ConsortiumStep) CreateConsortium(centralTenant string, accessToken string, consortiumName string) string {
	consortiumMap := cs.GetConsortiumByName(true, centralTenant, accessToken, consortiumName)
	if consortiumMap != nil {
		consortiumId := consortiumMap.(map[string]any)["id"].(string)

		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Consortium %s is already created", consortiumName))

		return consortiumId
	}

	consortiumId := uuid.New()

	bytes, err := json.Marshal(map[string]any{"id": consortiumId, "name": consortiumName})
	if err != nil {
		slog.Error(cs.Action.Name, "error", err)
		panic(err)
	}

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	cs.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, "/consortia"), true, bytes, headers)

	slog.Info(cs.Action.Name, "text", fmt.Sprintf("Created %s consortium", consortiumName))

	return consortiumId.String()
}

func (cs *ConsortiumStep) GetConsortiumByName(panicOnError bool, centralTenant string, accessToken string, consortiumName string) any {
	requestURL := fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, fmt.Sprintf("/consortia?query=name==%s", consortiumName))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundConsortiumsMap := cs.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if foundConsortiumsMap["consortia"] == nil || len(foundConsortiumsMap["consortia"].([]any)) == 0 {
		return nil
	}

	return foundConsortiumsMap["consortia"].([]any)[0]
}

// ######## Consortium Tenants ########

func (cs *ConsortiumStep) GetSortedConsortiumTenants(consortiumName string, tenants map[string]any) ConsortiumTenants {
	var consortiumTenants ConsortiumTenants
	for tenant, properties := range tenants {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		if properties == nil {
			consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: 0})
			continue
		}

		isCentral := cs.getSortableIsCentral(properties.(map[string]any))
		consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: isCentral})
	}

	sort.Slice(consortiumTenants, func(i, j int) bool {
		return consortiumTenants[i].IsCentral > consortiumTenants[j].IsCentral
	})

	return consortiumTenants
}

func (cs *ConsortiumStep) GetConsortiumUsers(consortiumName string, users map[string]any) map[string]any {
	consortiumUsers := make(map[string]any)
	for username, properties := range users {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		consortiumUsers[username] = properties
	}

	return consortiumUsers
}

func (cs *ConsortiumStep) GetAdminUsername(centralTenant string, consortiumUsers map[string]any) string {
	for username, properties := range consortiumUsers {
		tenant := properties.(map[string]any)[field.UsersTenantEntry]
		if tenant != nil && tenant.(string) == centralTenant {
			return username
		}
	}

	return ""
}

func (cs *ConsortiumStep) CreateConsortiumTenants(centralTenant string, accessToken string, consortiumId string, consortiumTenants ConsortiumTenants, adminUsername string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	for _, consortiumTenant := range consortiumTenants {
		bytes, err := json.Marshal(map[string]any{
			"id":        consortiumTenant.Tenant,
			"code":      consortiumTenant.Tenant[0:3],
			"name":      consortiumTenant.Tenant,
			"isCentral": consortiumTenant.IsCentral,
		})
		if err != nil {
			slog.Error(cs.Action.Name, "error", err)
			panic(err)
		}

		existingTenant := cs.getConsortiumTenantByIdAndName(true, centralTenant, accessToken, consortiumId, consortiumTenant.Tenant)
		if existingTenant != nil {
			slog.Info(cs.Action.Name, "text", fmt.Sprintf("Consortium tenant %s is already created", consortiumTenant.Tenant))
			continue
		}

		var requestURL = fmt.Sprintf("/consortia/%s/tenants", consortiumId)
		if consortiumTenant.IsCentral == 0 {
			user := cs.UserStep.GetUser(true, centralTenant, accessToken, adminUsername)

			requestURL = fmt.Sprintf("/consortia/%s/tenants?adminUserId=%s", consortiumId, user.(map[string]any)["id"].(string))
		}

		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Trying to create %s consortium tenant for %s consortium", consortiumTenant.Tenant, consortiumId))

		cs.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, requestURL), true, bytes, headers)

		cs.checkConsortiumTenantStatus(centralTenant, consortiumId, consortiumTenant.Tenant, headers)
	}
}

func (cs *ConsortiumStep) getConsortiumTenantByIdAndName(panicOnError bool, centralTenant string, accessToken string, consortiumId string, tenant string) any {
	requestURL := fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, fmt.Sprintf("/consortia/%s/tenants", consortiumId))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundConsortiumTenantsMap := cs.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
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

func (cs *ConsortiumStep) checkConsortiumTenantStatus(centralTenant string, consortiumId string, tenant string, headers map[string]string) {
	requestURL := fmt.Sprintf("/consortia/%s/tenants/%s", consortiumId, tenant)

	foundConsortiumTenantMap := cs.HTTPClient.DoGetDecodeReturnMapStringAny(fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, requestURL), true, headers)
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
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("waiting for %s consortium tenant creation", tenant))
		time.Sleep(WaitConsortiumTenant)
		cs.checkConsortiumTenantStatus(centralTenant, consortiumId, tenant, headers)
		return
	case FAILED:
	case COMPLETED_WITH_ERRORS:
		helpers.LogErrorPanic(cs.Action, fmt.Errorf("%s consortium tenant not is created", tenant))
		return
	case COMPLETED:
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Created %s consortium tenant (%t) for %s consortium", tenant, foundConsortiumTenantMap["isCentral"], consortiumId))
		return
	}
}

// ######## Central Ordering ########

func (cs *ConsortiumStep) EnableCentralOrdering(centralTenant string, accessToken string) {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"

	enableCentralOrdering := cs.getEnableCentralOrderingByKey(true, centralTenant, accessToken, centralOrderingLookupKey)
	if enableCentralOrdering {
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("central ordering for %s tenant is already enabled", centralTenant))
		return
	}

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	bytes, err := json.Marshal(map[string]any{"key": centralOrderingLookupKey, "value": "true"})
	if err != nil {
		slog.Error(cs.Action.Name, "error", err)
		panic(err)
	}

	cs.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, "/orders-storage/settings"), true, bytes, headers)

	slog.Info(cs.Action.Name, "text", fmt.Sprintf("Enabled central ordering for %s tenant", centralTenant))
}

func (cs *ConsortiumStep) getEnableCentralOrderingByKey(panicOnError bool, centralTenant string, accessToken string, key string) bool {
	requestURL := fmt.Sprintf(helpers.GetGatewayURL(cs.Action), constant.KongPort, fmt.Sprintf("/orders-storage/settings?query=key==%s", key))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundSettingsMap := cs.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if foundSettingsMap["settings"] == nil || len(foundSettingsMap["settings"].([]any)) == 0 {
		return false
	}

	settings := foundSettingsMap["settings"].([]any)[0]
	value := settings.(map[string]any)["value"].(string)

	enableCentralOrdering, err := strconv.ParseBool(value)
	if err != nil {
		slog.Error(cs.Action.Name, "error", err)
		panic(err)
	}

	return enableCentralOrdering
}
