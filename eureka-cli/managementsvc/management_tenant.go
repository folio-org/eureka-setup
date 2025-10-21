package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/viper"
)

func (ms *ManagementSvc) GetTenants(consortiumName string, tenantType constant.TenantType) ([]any, error) {
	requestURL := ms.Action.CreateURL(constant.KongPort, "/tenants")
	if tenantType != constant.All {
		requestURL += fmt.Sprintf("?query=description==%s-%s", consortiumName, tenantType)
	}

	foundTenantsMap, err := ms.HTTPClient.GetDecodeReturnMapStringAny(requestURL, map[string]string{})
	if err != nil {
		return nil, err
	}

	if foundTenantsMap["tenants"] == nil || len(foundTenantsMap["tenants"].([]any)) == 0 {
		return nil, nil
	}

	return foundTenantsMap["tenants"].([]any), nil
}

func (ms *ManagementSvc) CreateTenants() error {
	requestURL := ms.Action.CreateURL(constant.KongPort, "/tenants")
	foundTenants := viper.GetStringMap(field.Tenants)

	for tenant, properties := range foundTenants {
		mapEntry := properties.(map[string]any)

		description := fmt.Sprintf("%s-%s", constant.NoneConsortium, constant.Default)

		consortiumName := helpers.GetAnyOrDefault(mapEntry, field.TenantsConsortiumEntry, nil)
		if consortiumName != nil {
			tenantType := constant.Member
			if helpers.GetBool(mapEntry, field.TenantsCentralTenantEntry) {
				tenantType = constant.Central
			}

			description = fmt.Sprintf("%s-%s", consortiumName, tenantType)
		}

		b, err := json.Marshal(map[string]string{"name": tenant, "description": description})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.PostReturnNoContent(requestURL, b, map[string]string{})
		if err != nil {
			return err
		}

		slog.Info(ms.Action.Name, "text", "Created tenant with description", "tenant", tenant, "description", description)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	foundTenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range foundTenants {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(viper.GetStringMap(field.Tenants)), tenant) {
			continue
		}

		requestURL := ms.Action.CreateURL(constant.KongPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", mapEntry["id"].(string)))

		_ = ms.HTTPClient.Delete(requestURL, map[string]string{})

		slog.Info(ms.Action.Name, "text", "Removed tenant", "tenant", tenant)
	}

	return nil
}
