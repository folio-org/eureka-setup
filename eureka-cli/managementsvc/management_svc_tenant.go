package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	tenantsvc "github.com/folio-org/eureka-cli/tenantsvc"
	"github.com/spf13/viper"
)

type ManagementSvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	TenantSvc  *tenantsvc.TenantSvc
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, tenantSvc *tenantsvc.TenantSvc) *ManagementSvc {
	return &ManagementSvc{
		Action:     action,
		HTTPClient: httpClient,
		TenantSvc:  tenantSvc,
	}
}

func (ms *ManagementSvc) GetTenants(consortiumName string, tenantType constant.TenantType) ([]any, error) {
	requestURL := ms.Action.CreateURL(constant.KongPort, "/tenants")
	if tenantType != constant.All {
		requestURL += fmt.Sprintf("?query=description==%s-%s", consortiumName, tenantType)
	}

	tt, err := ms.HTTPClient.GetDecodeReturnMapStringAny(requestURL, map[string]string{})
	if err != nil {
		return nil, err
	}

	if tt["tenants"] == nil || len(tt["tenants"].([]any)) == 0 {
		slog.Warn(ms.Action.Name, "text", "Did not find any tenants", "consortiumName", consortiumName, "tenantType", tenantType)
		return nil, nil
	}

	return tt["tenants"].([]any), nil
}

func (ms *ManagementSvc) CreateTenants() error {
	requestURL := ms.Action.CreateURL(constant.KongPort, "/tenants")
	tt := viper.GetStringMap(field.Tenants)

	for tenant, properties := range tt {
		mapEntry := properties.(map[string]any)

		description := fmt.Sprintf("%s-%s", constant.NoneConsortium, constant.Default)

		consortiumName := helpers.GetAnyOrDefault(mapEntry, field.TenantsConsortiumEntry, nil)
		if consortiumName != nil {
			tenantType := constant.Member
			if helpers.GetBool(mapEntry, field.TenantsCentralTenantEntry) {
				tenantType = constant.Central
			}

			description = fmt.Sprintf("%s-%s", consortiumName, tenantType)
		}

		bb, err := json.Marshal(map[string]string{
			"name":        tenant,
			"description": description,
		})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.PostReturnNoContent(requestURL, bb, map[string]string{})
		if err != nil {
			return err
		}

		slog.Info(ms.Action.Name, "text", "Created tenant with description", "tenant", tenant, "description", description)
	}

	return nil
}

func (ms *ManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	tt, err := ms.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tt {
		mapEntry := value.(map[string]any)

		tenant := mapEntry["name"].(string)

		if !slices.Contains(helpers.ConvertMapKeysToSlice(viper.GetStringMap(field.Tenants)), tenant) {
			continue
		}

		requestURL := ms.Action.CreateURL(constant.KongPort, fmt.Sprintf("/tenants/%s?purgeKafkaTopics=true", mapEntry["id"].(string)))

		err = ms.HTTPClient.Delete(requestURL, map[string]string{})
		if err != nil {
			return err
		}

		slog.Info(ms.Action.Name, "text", "Removed tenant", "tenant", tenant)
	}

	return nil
}
