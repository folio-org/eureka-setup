package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// ManagementTenantEntitlementManager defines the interface for tenant entitlement management operations
type ManagementTenantEntitlementManager interface {
	GetTenantEntitlements(tenantName string, includeModules bool) (models.TenantEntitlementResponse, error)
	CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error
	UpgradeTenantEntitlement(consortiumName string, tenantType constant.TenantType, newApplicationID string) error
	RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType, purgeSchemas bool) error
}

func (ms *ManagementSvc) GetTenantEntitlements(tenantName string, includeModules bool) (models.TenantEntitlementResponse, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/entitlements?tenant=%s&includeModules=%t", tenantName, includeModules))
	headers, err := helpers.SecureOkapiApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return models.TenantEntitlementResponse{}, err
	}

	var response models.TenantEntitlementResponse
	if err := ms.HTTPClient.GetReturnStruct(requestURL, headers, &response); err != nil {
		return models.TenantEntitlementResponse{}, err
	}

	return response, nil
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
		tenantName := helpers.GetString(entry, "name")
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"tenantId":     helpers.GetString(entry, "id"),
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

func (ms *ManagementSvc) UpgradeTenantEntitlement(consortiumName string, tenantType constant.TenantType, newApplicationID string) error {
	tenantParameters, err := ms.TenantSvc.GetEntitlementTenantParameters(consortiumName)
	if err != nil {
		return err
	}

	tenants, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return nil
	}

	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/entitlements?async=false&tenantParameters=%s", tenantParameters))
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return nil
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := helpers.GetString(entry, "name")
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"tenantId":     helpers.GetString(entry, "id"),
			"applications": []string{newApplicationID},
		})
		if err != nil {
			return err
		}

		var decodedResponse models.TenantEntitlementResponse
		if err := ms.HTTPClient.PutReturnStruct(requestURL, payload, headers, &decodedResponse); err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Upgraded tenant entitlement", "tenant", tenantName, "flowId", decodedResponse.FlowID)
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
		tenantName := helpers.GetString(entry, "name")
		if !helpers.HasTenant(tenantName, ms.Action.ConfigTenants) {
			continue
		}
		tenantID := helpers.GetString(entry, "id")

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
