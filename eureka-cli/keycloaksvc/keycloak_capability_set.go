package keycloaksvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (ks *KeycloakSvc) GetCapabilitySets(headers map[string]string) ([]any, error) {
	var foundCapabilitySets []any

	applications, err := ks.ManagementSvc.GetApplications()
	if err != nil {
		return nil, err
	}

	for _, value := range applications.ApplicationDescriptors {
		applicationID := value["id"].(string)

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationID))

		foundCapabilitySetsMap, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
		if err != nil {
			return nil, err
		}

		if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
			return nil, nil
		}

		foundCapabilitySets = append(foundCapabilitySets, foundCapabilitySetsMap["capabilitySets"].([]any)...)
	}

	return foundCapabilitySets, nil
}

func (ks *KeycloakSvc) GetCapabilitySetsByName(headers map[string]string, capabilitySetName string) ([]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilitySetName))

	foundCapabilitySetsMap, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
		return nil, nil
	}

	return foundCapabilitySetsMap["capabilitySets"].([]any), nil
}

func (ks *KeycloakSvc) AttachCapabilitySetsToRoles(tenant string, accessToken string) error {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/roles/capability-sets")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(field.Roles)

	foundRoles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	if len(foundRoles) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot attach capability sets, found no roles in tenant", "tenant", tenant)
		return nil
	}

	for _, roleValue := range foundRoles {
		mapEntry := roleValue.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		rolesMapConfig := rolesMap[roleName].(map[string]any)
		if tenant != rolesMapConfig[field.RolesTenantEntry].(string) {
			continue
		}

		capabilitySetIDs, err := ks.populateCapabilitySets(headers, rolesMapConfig[field.RolesCapabilitySetsEntry].([]any))
		if err != nil {
			return err
		}

		if len(capabilitySetIDs) == 0 {
			slog.Info(ks.Action.Name, "text", "No capability sets were attached to role in tenant", "role", roleName, "tenant", tenant)
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySetIDs); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySetIDs))
			batchCapabilitySetIDs := capabilitySetIDs[lowerBound:upperBound]

			slog.Info(ks.Action.Name, "text", "Attaching capability sets to role in tenant", "rangeStart", lowerBound, "rangeEnd", upperBound, "total", len(capabilitySetIDs), "role", roleName, "tenant", tenant)

			b, err := json.Marshal(map[string]any{
				"roleId":           mapEntry["id"].(string),
				"capabilitySetIds": batchCapabilitySetIDs,
			})
			if err != nil {
				return err
			}

			err = ks.HTTPClient.PostRetryReturnNoContent(requestURL, b, headers)
			if err != nil {
				return err
			}
		}

		slog.Info(ks.Action.Name, "text", "Attached capability sets to role in tenant", "count", len(capabilitySetIDs), "role", roleName, "tenant", tenant)
	}

	return nil
}

func (ks *KeycloakSvc) populateCapabilitySets(headers map[string]string, capabilitySetNames []any) ([]string, error) {
	var capabilitySets = []string{}
	if len(capabilitySetNames) == 0 {
		return capabilitySets, nil
	}

	if len(capabilitySetNames) == 1 && !slices.Contains(capabilitySetNames, "all") {
		for _, capabilitySetName := range capabilitySetNames {
			foundCapabilitySets, err := ks.GetCapabilitySetsByName(headers, capabilitySetName.(string))
			if err != nil {
				return nil, err
			}

			for _, value := range foundCapabilitySets {
				capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
			}

		}

		return capabilitySets, nil
	}

	foundCapabilitySets, err := ks.GetCapabilitySets(headers)
	if err != nil {
		return nil, err
	}

	for _, value := range foundCapabilitySets {
		capabilitySets = append(capabilitySets, value.(map[string]any)["id"].(string))
	}

	return capabilitySets, nil
}

func (ks *KeycloakSvc) DetachCapabilitySetsFromRoles(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	rolesMap := viper.GetStringMap(field.Roles)

	foundRoles, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	if len(foundRoles) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot detach capability sets, found no roles in tenant", "tenant", tenant)
		return nil
	}

	for _, value := range foundRoles {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		_ = ks.HTTPClient.Delete(requestURL, headers)

		slog.Info(ks.Action.Name, "text", "Detached capability sets from role in tenant", "role", roleName, "tenant", tenant)
	}

	return nil
}
