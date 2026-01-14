package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
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
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return nil, err
	}

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
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	for tenantName, properties := range ms.Action.ConfigTenants {
		entry := properties.(map[string]any)
		payload, err := json.Marshal(map[string]string{
			"name":        tenantName,
			"description": ms.GetTenantType(entry),
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

func (ms *ManagementSvc) GetTenantType(entry map[string]any) string {
	consortiumName := helpers.GetString(entry, field.TenantsConsortiumEntry)
	if consortiumName == "" {
		return fmt.Sprintf("%s-%s", constant.NoneConsortium, constant.Default)
	}

	tenantType := constant.Member
	if helpers.GetBool(entry, field.TenantsCentralTenantEntry) {
		tenantType = constant.Central
	}

	return fmt.Sprintf("%s-%s", consortiumName, tenantType)
}

func (ms *ManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := helpers.GetString(entry, "name")
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", helpers.GetString(entry, "id")))
		if err := ms.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Removed tenant", "tenant", tenantName, "tenantType", tenantType)
	}

	return nil
}
