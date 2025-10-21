package consortiumsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/google/uuid"
)

func (cs *ConsortiumSvc) GetConsortiumCentralTenant(consortiumName string, tenants map[string]any) string {
	for tenant, properties := range tenants {
		if properties == nil || !cs.isValidConsortium(consortiumName, properties) || cs.getSortableIsCentral(properties.(map[string]any)) == 0 {
			continue
		}

		return tenant
	}

	return ""
}

func (cs *ConsortiumSvc) getSortableIsCentral(mapEntry map[string]any) int {
	if helpers.GetBool(mapEntry, field.TenantsCentralTenantEntry) {
		return 1
	}

	return 0
}

func (cs *ConsortiumSvc) isValidConsortium(consortiumName string, properties any) bool {
	return properties.(map[string]any)[field.TenantsConsortiumEntry] == consortiumName
}

func (cs *ConsortiumSvc) CreateConsortium(centralTenant string, accessToken string, consortiumName string) (string, error) {
	consortiumMap, err := cs.GetConsortiumByName(centralTenant, accessToken, consortiumName)
	if err != nil {
		return "", err
	}

	if consortiumMap != nil {
		consortiumID := consortiumMap.(map[string]any)["id"].(string)

		slog.Info(cs.Action.Name, "text", "Consortium is already created", "consortium", consortiumName)

		return consortiumID, nil
	}

	consortiumID := uuid.New()

	b, err := json.Marshal(map[string]any{
		"id":   consortiumID,
		"name": consortiumName,
	})
	if err != nil {
		return "", err
	}

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	err = cs.HTTPClient.PostReturnNoContent(cs.Action.CreateURL(constant.KongPort, "/consortia"), b, headers)
	if err != nil {
		return "", err
	}

	slog.Info(cs.Action.Name, "text", "Created consortium", "consortium", consortiumName)

	return consortiumID.String(), nil
}

func (cs *ConsortiumSvc) GetConsortiumByName(centralTenant string, accessToken string, consortiumName string) (any, error) {
	requestURL := cs.Action.CreateURL(constant.KongPort, fmt.Sprintf("/consortia?query=name==%s", consortiumName))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundConsortiumsMap, err := cs.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if foundConsortiumsMap["consortia"] == nil || len(foundConsortiumsMap["consortia"].([]any)) == 0 {
		return nil, err
	}

	return foundConsortiumsMap["consortia"].([]any)[0], nil
}
