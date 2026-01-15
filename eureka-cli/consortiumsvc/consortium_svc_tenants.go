package consortiumsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// ConsortiumTenantHandler defines the interface for consortium tenant operations
type ConsortiumTenantHandler interface {
	GetSortedConsortiumTenants(consortiumName string) models.SortedConsortiumTenants
	CreateConsortiumTenants(centralTenant string, consortiumID string, consortiumTenants models.SortedConsortiumTenants, adminUsername string) error
}

func (cs *ConsortiumSvc) GetSortedConsortiumTenants(consortiumName string) models.SortedConsortiumTenants {
	var consortiumTenants models.SortedConsortiumTenants
	for tenantName, properties := range cs.Action.ConfigTenants {
		if properties == nil {
			continue
		}
		if !cs.isValidConsortium(consortiumName, properties) {
			continue
		}

		entry := properties.(map[string]any)
		isCentral := cs.getSortableIsCentral(entry)
		consortiumTenants = append(consortiumTenants, &models.SortedConsortiumTenant{
			Name:      tenantName,
			IsCentral: isCentral,
		})
	}
	sort.Slice(consortiumTenants, func(i, j int) bool {
		return consortiumTenants[i].IsCentral > consortiumTenants[j].IsCentral
	})

	return consortiumTenants
}

func (cs *ConsortiumSvc) CreateConsortiumTenants(centralTenant string, consortiumID string, consortiumTenants models.SortedConsortiumTenants, adminUsername string) error {
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}
	for _, consortiumTenant := range consortiumTenants {
		payload, err := json.Marshal(map[string]any{
			"id":        consortiumTenant.Name,
			"code":      consortiumTenant.Name[0:3],
			"name":      consortiumTenant.Name,
			"isCentral": consortiumTenant.IsCentral,
		})
		if err != nil {
			return err
		}

		existingTenant, err := cs.getConsortiumTenantByIDAndName(centralTenant, consortiumID, consortiumTenant.Name)
		if err != nil {
			return err
		}
		if existingTenant != nil {
			slog.Info(cs.Action.Name, "text", "Consortium tenant is already created", "tenant", consortiumTenant.Name)
			continue
		}

		var requestURL = fmt.Sprintf("/consortia/%s/tenants", consortiumID)
		if consortiumTenant.IsCentral == 0 {
			user, err := cs.UserSvc.Get(centralTenant, adminUsername)
			if err != nil {
				return err
			}

			requestURL = fmt.Sprintf("/consortia/%s/tenants?adminUserId=%s", consortiumID, user.ID)
		}

		slog.Info(cs.Action.Name, "text", "Trying to create consortium tenant", "tenant", consortiumTenant.Name, "consortium", consortiumID)
		finalRequestURL := cs.Action.GetRequestURL(constant.KongPort, requestURL)
		if err := cs.HTTPClient.PostReturnNoContent(finalRequestURL, payload, headers); err != nil {
			return err
		}
		if err := cs.checkConsortiumTenantStatus(centralTenant, consortiumID, consortiumTenant.Name, headers); err != nil {
			return err
		}
	}

	return nil
}

func (cs *ConsortiumSvc) getConsortiumTenantByIDAndName(centralTenant string, consortiumID string, tenant string) (any, error) {
	requestURL := cs.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/consortia/%s/tenants", consortiumID))
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)
	if err != nil {
		return nil, err
	}

	var decodedResponse models.ConsortiumTenantsResponse
	if err := cs.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}
	if len(decodedResponse.Tenants) == 0 {
		return nil, nil
	}
	for _, t := range decodedResponse.Tenants {
		if t.Name == tenant {
			return t.Name, nil
		}
	}

	return nil, nil
}

func (cs *ConsortiumSvc) checkConsortiumTenantStatus(centralTenant string, consortiumID string, tenantName string, headers map[string]string) error {
	requestURL := cs.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/consortia/%s/tenants/%s", consortiumID, tenantName))

	var decodedResponse models.ConsortiumTenantStatus
	if err := cs.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return err
	}

	const (
		IN_PROGRESS           string = "IN_PROGRESS"
		FAILED                string = "FAILED"
		COMPLETED             string = "COMPLETED"
		COMPLETED_WITH_ERRORS string = "COMPLETED_WITH_ERRORS"
	)
	switch decodedResponse.SetupStatus {
	case IN_PROGRESS:
		slog.Warn(cs.Action.Name, "text", "Waiting for consortium tenant creation", "tenant", tenantName)
		time.Sleep(constant.ConsortiumTenantStatusWait)
		if err := cs.checkConsortiumTenantStatus(centralTenant, consortiumID, tenantName, headers); err != nil {
			return err
		}

		return nil
	case FAILED:
	case COMPLETED_WITH_ERRORS:
		return errors.TenantNotCreated(tenantName)
	case COMPLETED:
		slog.Info(cs.Action.Name, "text", "Created consortium tenant", "tenant", tenantName, "isCentral", decodedResponse.IsCentral)
		return nil
	}

	return nil
}
