package keycloakstep

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

func (ks *KeycloakStep) GetCapabilitySets(headers map[string]string) ([]any, error) {
	var foundCapabilitySets []any

	applications, err := ks.ManagementStep.GetApplications()
	if err != nil {
		return nil, err
	}

	for _, value := range applications.ApplicationDescriptors {
		applicationId := value["id"].(string)

		requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationId))

		foundCapabilitySetsMap, err := ks.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
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

func (ks *KeycloakStep) GetCapabilitySetsByName(headers map[string]string, capabilitySetName string) ([]any, error) {
	requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", capabilitySetName))

	foundCapabilitySetsMap, err := ks.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundCapabilitySetsMap["capabilitySets"] == nil || len(foundCapabilitySetsMap["capabilitySets"].([]any)) == 0 {
		return nil, nil
	}

	return foundCapabilitySetsMap["capabilitySets"].([]any), nil
}

func (ks *KeycloakStep) AttachCapabilitySetsToRoles(tenant string, accessToken string) error {
	requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, "/roles/capability-sets")

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
		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Cannot attach capability sets, found no roles in %s tenant (realm)", tenant))
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

		capabilitySetIds, err := ks.populateCapabilitySets(headers, rolesMapConfig[field.RolesCapabilitySetsEntry].([]any))
		if err != nil {
			return err
		}

		if len(capabilitySetIds) == 0 {
			slog.Info(ks.Action.Name, "text", fmt.Sprintf("No capability sets were attached to %s role in %s tenant (realm)", roleName, tenant))
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(capabilitySetIds); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(capabilitySetIds))
			batchCapabilitySetIds := capabilitySetIds[lowerBound:upperBound]

			slog.Info(ks.Action.Name, "text", fmt.Sprintf("Attaching %d-%d (total: %d) capability sets to %s role in %s tenant (realm)", lowerBound, upperBound, len(capabilitySetIds), roleName, tenant))

			b, err := json.Marshal(map[string]any{"roleId": mapEntry["id"].(string), "capabilitySetIds": batchCapabilitySetIds})
			if err != nil {
				return err
			}

			err = ks.HTTPClient.DoRetryPostReturnNoContent(requestURL, b, headers)
			if err != nil {
				return err
			}
		}

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Attached %d capability sets to %s role in %s tenant (realm)", len(capabilitySetIds), roleName, tenant))
	}

	return nil
}

func (ks *KeycloakStep) populateCapabilitySets(headers map[string]string, capabilitySetNames []any) ([]string, error) {
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

func (ks *KeycloakStep) DetachCapabilitySetsFromRoles(tenant string, accessToken string) error {
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
		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Cannot detach capability sets, found no roles in %s tenant (realm)", tenant))
		return nil
	}

	for _, value := range foundRoles {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rolesMap[roleName] == nil {
			continue
		}

		requestURL := fmt.Sprintf(ks.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		_ = ks.HTTPClient.DoDelete(requestURL, headers)

		slog.Info(ks.Action.Name, "text", fmt.Sprintf("Detached capability sets from %s role in %s tenant (realm)", roleName, tenant))
	}

	return nil
}
