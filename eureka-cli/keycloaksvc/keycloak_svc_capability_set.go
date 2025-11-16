package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// KeycloakCapabilitySetManager defines the interface for Keycloak capability set management operations
type KeycloakCapabilitySetManager interface {
	GetCapabilitySets(headers map[string]string) ([]any, error)
	GetCapabilitySetsByName(headers map[string]string, capabilityName string) ([]any, error)
	AttachCapabilitySetsToRoles(tenantName string) error
	DetachCapabilitySetsFromRoles(tenantName string) error
}

func (ks *KeycloakSvc) GetCapabilitySets(headers map[string]string) ([]any, error) {
	var capabilitySets []any
	applications, err := ks.ManagementSvc.GetApplications()
	if err != nil {
		return nil, err
	}

	for _, descriptor := range applications.ApplicationDescriptors {
		applicationID := descriptor["id"].(string)
		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/capability-sets?query=applicationId==%s&offset=0&limit=10000", applicationID))

		var decodedResponse models.KeycloakCapabilitySetsResponse
		if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
			return nil, err
		}
		if len(decodedResponse.CapabilitySets) == 0 {
			continue
		}
		for _, cs := range decodedResponse.CapabilitySets {
			capabilitySets = append(capabilitySets, map[string]any{
				"id":            cs.ID,
				"name":          cs.Name,
				"description":   cs.Description,
				"applicationId": cs.ApplicationID,
				"resource":      cs.Resource,
				"action":        cs.Action,
			})
		}
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) GetCapabilitySetsByName(headers map[string]string, capabilityName string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/capability-sets?query=name==%s&limit=1", capabilityName))

	var decodedResponse models.KeycloakCapabilitySetsResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}
	if len(decodedResponse.CapabilitySets) == 0 {
		return nil, nil
	}

	result := make([]any, len(decodedResponse.CapabilitySets))
	for i, cs := range decodedResponse.CapabilitySets {
		result[i] = map[string]any{
			"id":            cs.ID,
			"name":          cs.Name,
			"description":   cs.Description,
			"applicationId": cs.ApplicationID,
			"resource":      cs.Resource,
			"action":        cs.Action,
		}
	}

	return result, nil
}

func (ks *KeycloakSvc) AttachCapabilitySetsToRoles(tenantName string) error {
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		slog.Warn(ks.Action.Name, "text", "Found no roles with capability sets", "tenant", tenantName)
		return nil
	}

	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles/capability-sets")
	for _, roleValue := range roles {
		entry := roleValue.(map[string]any)
		roleName := ks.Action.Caser.String(entry["name"].(string))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		rolesMapConfig := ks.Action.ConfigRoles[roleName].(map[string]any)
		if tenantName != rolesMapConfig[field.RolesTenantEntry].(string) {
			continue
		}

		rolesCapabilitySets := rolesMapConfig[field.RolesCapabilitySetsEntry].([]any)
		capabilitySets, err := ks.populateCapabilitySets(headers, rolesCapabilitySets)
		if err != nil {
			return err
		}
		if len(capabilitySets) == 0 {
			slog.Warn(ks.Action.Name, "text", "No capability sets were attached", "role", roleName, "tenant", tenantName)
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySets); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySets))
			batchCapabilitySetIDs := capabilitySets[lowerBound:upperBound]
			slog.Info(ks.Action.Name, "text", "Attaching capability sets", "start", lowerBound, "end", upperBound, "total", len(capabilitySets), "role", roleName, "tenant", tenantName)

			payload, err := json.Marshal(map[string]any{
				"roleId":           entry["id"].(string),
				"capabilitySetIds": batchCapabilitySetIDs,
			})
			if err != nil {
				return err
			}
			if err := ks.HTTPClient.PostRetryReturnNoContent(requestURL, payload, headers); err != nil {
				return err
			}
		}
		slog.Info(ks.Action.Name, "text", "Attached capability sets", "count", len(capabilitySets), "role", roleName, "tenant", tenantName)
	}

	return nil
}

func (ks *KeycloakSvc) populateCapabilitySets(headers map[string]string, rolesCapabilitySets []any) ([]string, error) {
	if len(rolesCapabilitySets) == 0 {
		return []string{}, nil
	}

	if len(rolesCapabilitySets) == 1 && !slices.Contains(rolesCapabilitySets, "all") {
		var capabilitySets = []string{}
		for _, capabilitySetName := range rolesCapabilitySets {
			capabilitySetsFound, err := ks.GetCapabilitySetsByName(headers, capabilitySetName.(string))
			if err != nil {
				return nil, err
			}
			for _, value := range capabilitySetsFound {
				capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
			}
		}
		return capabilitySets, nil
	}

	var capabilitySets = []string{}
	allCapabilitySets, err := ks.GetCapabilitySets(headers)
	if err != nil {
		return nil, err
	}
	for _, value := range allCapabilitySets {
		capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) DetachCapabilitySetsFromRoles(tenantName string) error {
	headers := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		slog.Warn(ks.Action.Name, "text", "Found no roles with capability sets", "tenant", tenantName)
		return nil
	}

	for _, value := range roles {
		entry := value.(map[string]any)
		roleName := ks.Action.Caser.String(entry["name"].(string))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", entry["id"].(string)))
		if err := ks.HTTPClient.Delete(requestURL, headers); err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Detached capability sets", "role", roleName, "tenant", tenantName)
	}

	return nil
}
