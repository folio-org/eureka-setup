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
	var cc1 []any

	applications, err := ks.ManagementSvc.GetApplications()
	if err != nil {
		return nil, err
	}

	for _, value := range applications.ApplicationDescriptors {
		applicationID := value["id"].(string)

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=10000&query=applicationId==%s", applicationID))

		cc2, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
		if err != nil {
			return nil, err
		}

		if cc2["capabilitySets"] == nil || len(cc2["capabilitySets"].([]any)) == 0 {
			return nil, nil
		}

		cc1 = append(cc1, cc2["capabilitySets"].([]any)...)
	}

	return cc1, nil
}

func (ks *KeycloakSvc) GetCapabilitySetsByName(headers map[string]string, cc1 string) ([]any, error) {
	requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/capability-sets?offset=0&limit=1000&query=name=%s", cc1))

	cc2, err := ks.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if cc2["capabilitySets"] == nil || len(cc2["capabilitySets"].([]any)) == 0 {
		return nil, nil
	}

	return cc2["capabilitySets"].([]any), nil
}

func (ks *KeycloakSvc) AttachCapabilitySetsToRoles(tenant string, accessToken string) error {
	requestURL := ks.Action.CreateURL(constant.KongPort, "/roles/capability-sets")

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	rr1 := viper.GetStringMap(field.Roles)

	rr2, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	if len(rr2) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot attach capability sets, found no roles in tenant", "tenant", tenant)
		return nil
	}

	for _, roleValue := range rr2 {
		mapEntry := roleValue.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rr1[roleName] == nil {
			continue
		}

		rolesMapConfig := rr1[roleName].(map[string]any)
		if tenant != rolesMapConfig[field.RolesTenantEntry].(string) {
			continue
		}

		cc, err := ks.populateCapabilitySets(headers, rolesMapConfig[field.RolesCapabilitySetsEntry].([]any))
		if err != nil {
			return err
		}

		if len(cc) == 0 {
			slog.Info(ks.Action.Name, "text", "No capability sets were attached to role in tenant", "role", roleName, "tenant", tenant)
			continue
		}

		batchSize := 250
		for lowerBound := 0; lowerBound < len(cc); lowerBound += batchSize {
			upperBound := min(lowerBound+batchSize, len(cc))
			batchCapabilitySetIDs := cc[lowerBound:upperBound]

			slog.Info(ks.Action.Name, "text", "Attaching capability sets to role in tenant", "rangeStart", lowerBound, "rangeEnd", upperBound, "total", len(cc), "role", roleName, "tenant", tenant)

			bb, err := json.Marshal(map[string]any{
				"roleId":           mapEntry["id"].(string),
				"capabilitySetIds": batchCapabilitySetIDs,
			})
			if err != nil {
				return err
			}

			err = ks.HTTPClient.PostRetryReturnNoContent(requestURL, bb, headers)
			if err != nil {
				return err
			}
		}

		slog.Info(ks.Action.Name, "text", "Attached capability sets to role in tenant", "count", len(cc), "role", roleName, "tenant", tenant)
	}

	return nil
}

func (ks *KeycloakSvc) populateCapabilitySets(headers map[string]string, cc1 []any) ([]string, error) {
	if len(cc1) == 0 {
		return []string{}, nil
	}

	if len(cc1) == 1 && !slices.Contains(cc1, "all") {
		var cc2 = []string{}
		for _, capabilitySetName := range cc1 {
			cc3, err := ks.GetCapabilitySetsByName(headers, capabilitySetName.(string))
			if err != nil {
				return nil, err
			}

			for _, value := range cc3 {
				cc2 = append(cc2, value.(map[string]any)["id"].(string))
			}

		}

		return cc2, nil
	}

	var cc2 = []string{}
	cc3, err := ks.GetCapabilitySets(headers)
	if err != nil {
		return nil, err
	}

	for _, value := range cc3 {
		cc2 = append(cc2, value.(map[string]any)["id"].(string))
	}

	return cc2, nil
}

func (ks *KeycloakSvc) DetachCapabilitySetsFromRoles(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	caser := cases.Lower(language.English)
	rr1 := viper.GetStringMap(field.Roles)

	rr2, err := ks.GetRoles(headers)
	if err != nil {
		return err
	}

	if len(rr2) == 0 {
		slog.Info(ks.Action.Name, "text", "Cannot detach capability sets, found no roles in tenant", "tenant", tenant)
		return nil
	}

	for _, value := range rr2 {
		mapEntry := value.(map[string]any)

		roleName := caser.String(mapEntry["name"].(string))
		if rr1[roleName] == nil {
			continue
		}

		requestURL := ks.Action.CreateURL(constant.KongPort, fmt.Sprintf("/roles/%s/capability-sets", mapEntry["id"].(string)))

		err = ks.HTTPClient.Delete(requestURL, headers)
		if err != nil {
			return err
		}

		slog.Info(ks.Action.Name, "text", "Detached capability sets from role in tenant", "role", roleName, "tenant", tenant)
	}

	return nil
}
