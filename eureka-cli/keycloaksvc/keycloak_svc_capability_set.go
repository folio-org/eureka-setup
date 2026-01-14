package keycloaksvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	apperrors "github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
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
		applicationID := helpers.GetString(descriptor, "id")
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
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

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
		roleName := ks.Action.Caser.String(helpers.GetString(entry, "name"))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		rolesMapConfig := helpers.GetMapOrDefault(ks.Action.ConfigRoles, roleName, nil)
		if tenantName != helpers.GetString(rolesMapConfig, field.RolesTenantEntry) {
			continue
		}

		rolesCapabilitySets := helpers.GetAnySlice(rolesMapConfig, field.RolesCapabilitySetsEntry)
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
				"roleId":           helpers.GetString(entry, "id"),
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
				rawCapabilitySets := value.(map[string]any)
				capabilitySets = append(capabilitySets, helpers.GetString(rawCapabilitySets, "id"))
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
		rawCapabilitySets := value.(map[string]any)
		capabilitySets = append(capabilitySets, helpers.GetString(rawCapabilitySets, "id"))
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) DetachCapabilitySetsFromRoles(tenantName string) error {
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

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
		roleName := ks.Action.Caser.String(helpers.GetString(entry, "name"))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", helpers.GetString(entry, "id")))
		if err := ks.HTTPClient.Delete(requestURL, headers); err != nil {
			if errors.Is(err, apperrors.ErrHTTP404NotFound) {
				slog.Debug(ks.Action.Name, "text", "No capability sets to detach (already detached or not found)", "role", roleName, "tenant", tenantName)
				continue
			}
			return err
		}
		slog.Info(ks.Action.Name, "text", "Detached capability sets", "role", roleName, "tenant", tenantName)
	}

	return nil
}
