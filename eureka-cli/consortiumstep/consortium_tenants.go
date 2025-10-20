package consortiumstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
)

type ConsortiumTenant struct {
	Consortium string
	Tenant     string
	IsCentral  int
}

type ConsortiumTenants []*ConsortiumTenant

func (c ConsortiumTenant) String() string {
	if c.IsCentral == 1 {
		return fmt.Sprintf("%s (central)", c.Tenant)
	}

	return c.Tenant
}

func (c ConsortiumTenants) String() string {
	var builder strings.Builder
	for idx, value := range c {
		builder.WriteString(value.Tenant)
		if idx+1 < len(c) {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

func (cs *ConsortiumStep) GetSortedConsortiumTenants(consortiumName string, tenants map[string]any) ConsortiumTenants {
	var consortiumTenants ConsortiumTenants
	for tenant, properties := range tenants {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		if properties == nil {
			consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: 0})
			continue
		}

		isCentral := cs.getSortableIsCentral(properties.(map[string]any))
		consortiumTenants = append(consortiumTenants, &ConsortiumTenant{Tenant: tenant, IsCentral: isCentral})
	}

	sort.Slice(consortiumTenants, func(i, j int) bool {
		return consortiumTenants[i].IsCentral > consortiumTenants[j].IsCentral
	})

	return consortiumTenants
}

func (cs *ConsortiumStep) GetConsortiumUsers(consortiumName string, users map[string]any) map[string]any {
	consortiumUsers := make(map[string]any)
	for username, properties := range users {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		consortiumUsers[username] = properties
	}

	return consortiumUsers
}

func (cs *ConsortiumStep) GetAdminUsername(centralTenant string, consortiumUsers map[string]any) string {
	for username, properties := range consortiumUsers {
		tenant := properties.(map[string]any)[field.UsersTenantEntry]
		if tenant != nil && tenant.(string) == centralTenant {
			return username
		}
	}

	return ""
}

func (cs *ConsortiumStep) CreateConsortiumTenants(centralTenant string, accessToken string, consortiumId string, consortiumTenants ConsortiumTenants, adminUsername string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	for _, consortiumTenant := range consortiumTenants {
		b, err := json.Marshal(map[string]any{
			"id":        consortiumTenant.Tenant,
			"code":      consortiumTenant.Tenant[0:3],
			"name":      consortiumTenant.Tenant,
			"isCentral": consortiumTenant.IsCentral,
		})
		if err != nil {
			slog.Error(cs.Action.Name, "error", err)
			panic(err)
		}

		existingTenant, err := cs.getConsortiumTenantByIdAndName(centralTenant, accessToken, consortiumId, consortiumTenant.Tenant)
		if err != nil {
			return err
		}

		if existingTenant != nil {
			slog.Info(cs.Action.Name, "text", fmt.Sprintf("Consortium tenant %s is already created", consortiumTenant.Tenant))
			continue
		}

		var requestURL = fmt.Sprintf("/consortia/%s/tenants", consortiumId)
		if consortiumTenant.IsCentral == 0 {
			user, err := cs.UserStep.GetUser(centralTenant, accessToken, adminUsername)
			if err != nil {
				return err
			}

			requestURL = fmt.Sprintf("/consortia/%s/tenants?adminUserId=%s", consortiumId, user.(map[string]any)["id"].(string))
		}

		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Trying to create %s consortium tenant for %s consortium", consortiumTenant.Tenant, consortiumId))

		err = cs.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(cs.Action.GatewayURL, constant.KongPort, requestURL), b, headers)
		if err != nil {
			return err
		}

		cs.checkConsortiumTenantStatus(centralTenant, consortiumId, consortiumTenant.Tenant, headers)
	}

	return nil
}

func (cs *ConsortiumStep) getConsortiumTenantByIdAndName(centralTenant string, accessToken string, consortiumId string, tenant string) (any, error) {
	requestURL := fmt.Sprintf(cs.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/consortia/%s/tenants", consortiumId))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundConsortiumTenantsMap, err := cs.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundConsortiumTenantsMap["tenants"] == nil || len(foundConsortiumTenantsMap["tenants"].([]any)) == 0 {
		return nil, nil
	}

	for _, value := range foundConsortiumTenantsMap["tenants"].([]any) {
		existingTenant := value.(map[string]any)["name"]
		if existingTenant != nil && existingTenant.(string) == tenant {
			return existingTenant, nil
		}
	}

	return nil, nil
}

func (cs *ConsortiumStep) checkConsortiumTenantStatus(centralTenant string, consortiumId string, tenant string, headers map[string]string) error {
	requestURL := fmt.Sprintf("/consortia/%s/tenants/%s", consortiumId, tenant)

	foundConsortiumTenantMap, err := cs.HTTPClient.DoGetDecodeReturnMapStringAny(fmt.Sprintf(cs.Action.GatewayURL, constant.KongPort, requestURL), headers)
	if err != nil {
		return err
	}

	if foundConsortiumTenantMap == nil {
		return nil
	}

	const (
		IN_PROGRESS           string = "IN_PROGRESS"
		FAILED                string = "FAILED"
		COMPLETED             string = "COMPLETED"
		COMPLETED_WITH_ERRORS string = "COMPLETED_WITH_ERRORS"

		WaitConsortiumTenant time.Duration = 10 * time.Second
	)

	switch foundConsortiumTenantMap["setupStatus"] {
	case IN_PROGRESS:
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("waiting for %s consortium tenant creation", tenant))
		time.Sleep(WaitConsortiumTenant)
		cs.checkConsortiumTenantStatus(centralTenant, consortiumId, tenant, headers)
		return nil
	case FAILED:
	case COMPLETED_WITH_ERRORS:
		return fmt.Errorf("%s consortium tenant not is created", tenant)
	case COMPLETED:
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Created %s consortium tenant (%t) for %s consortium", tenant, foundConsortiumTenantMap["isCentral"], consortiumId))
		return nil
	}

	return nil
}
