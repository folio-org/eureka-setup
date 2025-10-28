package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
)

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
		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationID))
		resp, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
		if err != nil {
			return nil, err
		}
		if resp["capabilitySets"] == nil || len(resp["capabilitySets"].([]any)) == 0 {
			return nil, nil
		}
		capabilitySets = append(capabilitySets, resp["capabilitySets"].([]any)...)
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) GetCapabilitySetsByName(headers map[string]string, capabilityName string) ([]any, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilityName))
	resp, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}
	if resp["capabilitySets"] == nil || len(resp["capabilitySets"].([]any)) == 0 {
		return nil, nil
	}

	return resp["capabilitySets"].([]any), nil
}

func (ks *KeycloakSvc) AttachCapabilitySetsToRoles(tenantName string) error {
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	requestURL := ks.Action.GetRequestURL(constant.KongPort, "/roles/capability-sets")
	if len(roles) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot attach capability sets, found no roles in tenant", "tenant", tenantName)
		return nil
	}

	for _, roleValue := range roles {
		mapEntry := roleValue.(map[string]any)
		roleName := ks.Action.Caser.String(mapEntry["name"].(string))
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
			slog.Info(ks.Action.Name, "text", "No capability sets were attached to role in tenant", "role", roleName, "tenant", tenantName)
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySets); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySets))
			batchCapabilitySetIDs := capabilitySets[lowerBound:upperBound]
			slog.Info(ks.Action.Name, "text", "Attaching capability sets to role in tenant", "rangeStart", lowerBound, "rangeEnd", upperBound, "total", len(capabilitySets), "role", roleName, "tenant", tenantName)

			payload, err := json.Marshal(map[string]any{
				"roleId":           mapEntry["id"].(string),
				"capabilitySetIds": batchCapabilitySetIDs,
			})
			if err != nil {
				return err
			}

			err = ks.HTTPClient.PostRetryReturnNoContent(requestURL, payload, headers)
			if err != nil {
				return err
			}
		}
		slog.Info(ks.Action.Name, "text", "Attached capability sets to role in tenant", "count", len(capabilitySets), "role", roleName, "tenant", tenantName)
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
			resp, err := ks.GetCapabilitySetsByName(headers, capabilitySetName.(string))
			if err != nil {
				return nil, err
			}
			for _, value := range resp {
				capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
			}
		}
		return capabilitySets, nil
	}

	var capabilitySets = []string{}
	resp, err := ks.GetCapabilitySets(headers)
	if err != nil {
		return nil, err
	}
	for _, value := range resp {
		capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) DetachCapabilitySetsFromRoles(tenantName string) error {
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ks.Action.KeycloakAccessToken)
	roles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot detach capability sets, found no roles in tenant", "tenant", tenantName)
		return nil
	}

	for _, value := range roles {
		mapEntry := value.(map[string]any)
		roleName := ks.Action.Caser.String(mapEntry["name"].(string))
		if ks.Action.ConfigRoles[roleName] == nil {
			continue
		}

		requestURL := ks.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))
		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}
		slog.Info(ks.Action.Name, "text", "Detached capability sets from role in tenant", "role", roleName, "tenant", tenantName)
	}

	return nil
}
