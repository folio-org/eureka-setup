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

func (ms *ManagementSvc) CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error {
	tenants := viper.GetStringMap(field.Tenants)

	tenantParameters, err := ms.TenantSvc.GetTenantParameters(consortiumName, tenants)
	if err != nil {
		return err
	}

	requestURL := ms.Action.CreateURL(constant.KongPort, fmt.Sprintf("/entitlements?purgeOnRollback=true&ignoreErrors=false&tenantParameters=%s", tenantParameters))
	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

	foundTenants, _ := ms.GetTenants(consortiumName, tenantType)

	for _, value := range foundTenants {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(tenants), tenant) {
			continue
		}

		applications := []string{fmt.Sprintf("%s-%s", applicationName, applicationVersion)}

		b, err := json.Marshal(map[string]any{
			"tenantId":     mapEntry["id"].(string),
			"applications": applications,
		})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.PostReturnNoContent(requestURL, b, map[string]string{})
		if err != nil {
			return err
		}

		slog.Info(ms.Action.Name, "text", "Created tenant entitlement for tenant", "tenant", tenant)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenantEntitlements(purgeSchemas bool, consortiumName string, tenantType constant.TenantType) error {
	requestURL := ms.Action.CreateURL(constant.KongPort, fmt.Sprintf("/entitlements?purge=%t&ignoreErrors=false", purgeSchemas))
	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)

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

		tenantID := mapEntry["id"].(string)

		var applications []string
		applications = append(applications, fmt.Sprintf("%s-%s", applicationName, applicationVersion))

		b, err := json.Marshal(map[string]any{
			"tenantId":     tenantID,
			"applications": applications,
		})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.DeleteWithBody(requestURL, b, map[string]string{})
		if err != nil {
			return err
		}

		slog.Info(ms.Action.Name, "text", "Removed tenant entitlement for tenant", "tenant", tenant)
	}

	return nil
}
