package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
)

// ManagementTenantManager defines the interface for tenant management operations
type ManagementTenantManager interface {
	GetTenants(consortiumName string, tenantType constant.TenantType) ([]any, error)
	CreateTenants() error
	RemoveTenants(consortiumName string, tenantType constant.TenantType) error
}

func (ms *ManagementSvc) GetTenants(consortiumName string, tenantType constant.TenantType) ([]any, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, "/tenants")
	if tenantType != constant.All {
		requestURL += fmt.Sprintf("?query=description==%s-%s", consortiumName, tenantType)
	}

	decodedResponse, err := ms.HTTPClient.GetRetryDecodeReturnAny(requestURL, map[string]string{})
	if err != nil {
		return nil, err
	}

	tenantsData := decodedResponse.(map[string]any)
	if tenantsData["tenants"] == nil || len(tenantsData["tenants"].([]any)) == 0 {
		slog.Warn(ms.Action.Name, "text", "Did not find any tenants", "consortium", consortiumName, "tenantType", tenantType)
		return nil, nil
	}

	return tenantsData["tenants"].([]any), nil
}

func (ms *ManagementSvc) CreateTenants() error {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, "/tenants")
	for tenantName, properties := range ms.Action.ConfigTenants {
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

		payload, err := json.Marshal(map[string]string{
			"name":        tenantName,
			"description": description,
		})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.PostReturnNoContent(requestURL, payload, map[string]string{})
		if err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Created tenant with description", "tenant", tenantName, "description", description)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		mapEntry := value.(map[string]any)
		tenantName := mapEntry["name"].(string)
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", mapEntry["id"].(string)))
		err = ms.HTTPClient.Delete(requestURL, map[string]string{})
		if err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Removed tenant", "tenant", tenantName)
	}

	return nil
}
