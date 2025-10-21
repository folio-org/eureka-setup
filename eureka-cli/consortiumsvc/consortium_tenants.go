package consortiumsvc

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

func (cs *ConsortiumSvc) GetSortedConsortiumTenants(consortiumName string, tenants map[string]any) ConsortiumTenants {
	var tt ConsortiumTenants
	for tenant, properties := range tenants {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		if properties == nil {
			tt = append(tt, &ConsortiumTenant{Tenant: tenant, IsCentral: 0})
			continue
		}

		isCentral := cs.getSortableIsCentral(properties.(map[string]any))
		tt = append(tt, &ConsortiumTenant{
			Tenant:    tenant,
			IsCentral: isCentral,
		})
	}

	sort.Slice(tt, func(i, j int) bool {
		return tt[i].IsCentral > tt[j].IsCentral
	})

	return tt
}

func (cs *ConsortiumSvc) GetConsortiumUsers(consortiumName string, users map[string]any) map[string]any {
	uu := make(map[string]any)
	for username, properties := range users {
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		uu[username] = properties
	}

	return uu
}

func (cs *ConsortiumSvc) GetAdminUsername(centralTenant string, uu map[string]any) string {
	for username, properties := range uu {
		tenant := properties.(map[string]any)[field.UsersTenantEntry]
		if tenant != nil && tenant.(string) == centralTenant {
			return username
		}
	}

	return ""
}

func (cs *ConsortiumSvc) CreateConsortiumTenants(centralTenant string, accessToken string, consortiumID string, consortiumTenants ConsortiumTenants, adminUsername string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	for _, consortiumTenant := range consortiumTenants {
		bb, err := json.Marshal(map[string]any{
			"id":        consortiumTenant.Tenant,
			"code":      consortiumTenant.Tenant[0:3],
			"name":      consortiumTenant.Tenant,
			"isCentral": consortiumTenant.IsCentral,
		})
		if err != nil {
			return err
		}

		existingTenant, err := cs.getConsortiumTenantByIDAndName(centralTenant, accessToken, consortiumID, consortiumTenant.Tenant)
		if err != nil {
			return err
		}

		if existingTenant != nil {
			slog.Info(cs.Action.Name, "text", "Consortium tenant is already created", "tenant", consortiumTenant.Tenant)
			continue
		}

		var requestURL = fmt.Sprintf("/consortia/%s/tenants", consortiumID)
		if consortiumTenant.IsCentral == 0 {
			user, err := cs.UserSvc.GetUser(centralTenant, accessToken, adminUsername)
			if err != nil {
				return err
			}

			requestURL = fmt.Sprintf("/consortia/%s/tenants?adminUserId=%s", consortiumID, user.(map[string]any)["id"].(string))
		}

		slog.Info(cs.Action.Name, "text", "Trying to create consortium tenant for consortium", "tenant", consortiumTenant.Tenant, "consortium", consortiumID)

		err = cs.HTTPClient.PostReturnNoContent(cs.Action.CreateURL(constant.KongPort, requestURL), bb, headers)
		if err != nil {
			return err
		}

		err = cs.checkConsortiumTenantStatus(centralTenant, consortiumID, consortiumTenant.Tenant, headers)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *ConsortiumSvc) getConsortiumTenantByIDAndName(centralTenant string, accessToken string, consortiumID string, tenant string) (any, error) {
	requestURL := cs.Action.CreateURL(constant.KongPort, fmt.Sprintf("/consortia/%s/tenants", consortiumID))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	tt, err := cs.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if tt["tenants"] == nil || len(tt["tenants"].([]any)) == 0 {
		return nil, nil
	}

	for _, value := range tt["tenants"].([]any) {
		existingTenant := value.(map[string]any)["name"]
		if existingTenant != nil && existingTenant.(string) == tenant {
			return existingTenant, nil
		}
	}

	return nil, nil
}

func (cs *ConsortiumSvc) checkConsortiumTenantStatus(centralTenant string, consortiumID string, tenant string, headers map[string]string) error {
	requestURL := fmt.Sprintf("/consortia/%s/tenants/%s", consortiumID, tenant)

	tt, err := cs.HTTPClient.GetDecodeReturnMapStringAny(cs.Action.CreateURL(constant.KongPort, requestURL), headers)
	if err != nil {
		return err
	}

	if tt == nil {
		return nil
	}

	const (
		IN_PROGRESS           string = "IN_PROGRESS"
		FAILED                string = "FAILED"
		COMPLETED             string = "COMPLETED"
		COMPLETED_WITH_ERRORS string = "COMPLETED_WITH_ERRORS"

		WaitConsortiumTenant time.Duration = 10 * time.Second
	)

	switch tt["setupStatus"] {
	case IN_PROGRESS:
		slog.Info(cs.Action.Name, "text", "Waiting for consortium tenant creation", "tenant", tenant)
		time.Sleep(WaitConsortiumTenant)
		err = cs.checkConsortiumTenantStatus(centralTenant, consortiumID, tenant, headers)
		if err != nil {
			return err
		}
		return nil
	case FAILED:
	case COMPLETED_WITH_ERRORS:
		return fmt.Errorf("%s consortium tenant not is created", tenant)
	case COMPLETED:
		slog.Info(cs.Action.Name, "text", "Created consortium tenant for consortium", "tenant", tenant, "isCentral", tt["isCentral"], "consortium", consortiumID)
		return nil
	}

	return nil
}
