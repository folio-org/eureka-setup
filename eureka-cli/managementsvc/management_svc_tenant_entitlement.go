package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// ManagementTenantEntitlementManager defines the interface for tenant entitlement management operations
type ManagementTenantEntitlementManager interface {
	CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error
	RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType, purgeSchemas bool) error
}

func (ms *ManagementSvc) CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error {
	tenantParameters, err := ms.TenantSvc.GetEntitlementTenantParameters(consortiumName)
	if err != nil {
		return err
	}

	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return nil
	}

	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/entitlements?purgeOnRollback=true&ignoreErrors=false&async=false&tenantParameters=%s", tenantParameters))
	headers, err := helpers.SecureOkapiApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := entry["name"].(string)
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"tenantId":     entry["id"].(string),
			"applications": []string{ms.Action.ConfigApplicationID},
		})
		if err != nil {
			return err
		}

		var decodedResponse models.TenantEntitlementResponse
		if err := ms.HTTPClient.PostReturnStruct(requestURL, payload, headers, &decodedResponse); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Created tenant entitlement", "tenant", tenantName, "flowId", decodedResponse.FlowID)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType, purgeSchemas bool) error {
	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/entitlements?purge=%t&ignoreErrors=false", purgeSchemas))
	headers, err := helpers.SecureOkapiApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := entry["name"].(string)
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}
		tenantID := entry["id"].(string)

		payload, err := json.Marshal(map[string]any{
			"tenantId":     tenantID,
			"applications": []string{ms.Action.ConfigApplicationID},
		})
		if err != nil {
			return err
		}

		var decodedResponse models.TenantEntitlementResponse
		if err := ms.HTTPClient.DeleteWithPayloadReturnStruct(requestURL, payload, headers, &decodedResponse); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Removed tenant entitlement", "tenant", tenantName, "flowId", decodedResponse.FlowID)
	}

	return nil
}
