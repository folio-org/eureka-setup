package consortiumstep

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/google/uuid"
)

func (cs *ConsortiumStep) GetConsortiumCentralTenant(consortiumName string, tenants map[string]any) string {
	for tenant, properties := range tenants {
		if properties == nil || !cs.isValidConsortium(consortiumName, properties) || cs.getSortableIsCentral(properties.(map[string]any)) == 0 {
			continue
		}

		return tenant
	}

	return ""
}

func (cs *ConsortiumStep) getSortableIsCentral(mapEntry map[string]any) int {
	if helpers.GetBool(mapEntry, field.TenantsCentralTenantEntry) {
		return 1
	}

	return 0
}

func (cs *ConsortiumStep) isValidConsortium(consortiumName string, properties any) bool {
	return properties.(map[string]any)[field.TenantsConsortiumEntry] == consortiumName
}

func (cs *ConsortiumStep) CreateConsortium(centralTenant string, accessToken string, consortiumName string) (string, error) {
	consortiumMap, err := cs.GetConsortiumByName(centralTenant, accessToken, consortiumName)
	if err != nil {
		return "", err
	}

	if consortiumMap != nil {
		consortiumId := consortiumMap.(map[string]any)["id"].(string)

		slog.Info(cs.Action.Name, "text", fmt.Sprintf("Consortium %s is already created", consortiumName))

		return consortiumId, nil
	}

	consortiumId := uuid.New()

	b, err := json.Marshal(map[string]any{"id": consortiumId, "name": consortiumName})
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

	slog.Info(cs.Action.Name, "text", fmt.Sprintf("Created %s consortium", consortiumName))

	return consortiumId.String(), nil
}

func (cs *ConsortiumStep) GetConsortiumByName(centralTenant string, accessToken string, consortiumName string) (any, error) {
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
