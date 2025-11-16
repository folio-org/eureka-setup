package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
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
	headers := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)

	var decodedResponse models.TenantsResponse
	if err := ms.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}

	if len(decodedResponse.Tenants) == 0 {
		slog.Warn(ms.Action.Name, "text", "Tenants are not found", "consortium", consortiumName, "tenantType", tenantType)
		return nil, nil
	}

	result := make([]any, len(decodedResponse.Tenants))
	for i, tenant := range decodedResponse.Tenants {
		result[i] = map[string]any{
			"id":          tenant.ID,
			"name":        tenant.Name,
			"description": tenant.Description,
		}
	}

	return result, nil
}

func (ms *ManagementSvc) CreateTenants() error {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, "/tenants")
	headers := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	for tenantName, properties := range ms.Action.ConfigTenants {
		entry := properties.(map[string]any)
		consortiumName := helpers.GetAnyOrDefault(entry, field.TenantsConsortiumEntry, nil)

		var description string
		if consortiumName == nil {
			description = fmt.Sprintf("%s-%s", constant.NoneConsortium, constant.Default)
		} else {
			tenantType := constant.Member
			if helpers.GetBool(entry, field.TenantsCentralTenantEntry) {
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

		var tenant models.Tenant
		if err := ms.HTTPClient.PostReturnStruct(requestURL, payload, headers, &tenant); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Created tenant", "tenant", tenant.Name, "id", tenant.ID, "description", tenant.Description)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	headers := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := entry["name"].(string)
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", entry["id"].(string)))
		if err := ms.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Removed tenant", "tenant", tenantName, "tenantType", tenantType)
	}

	return nil
}
