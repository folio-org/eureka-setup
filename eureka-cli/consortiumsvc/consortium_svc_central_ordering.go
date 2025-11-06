package consortiumsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// ConsortiumCentralOrderingManager defines the interface for consortium central ordering operations
type ConsortiumCentralOrderingManager interface {
	EnableCentralOrdering(centralTenant string) error
}

func (cs *ConsortiumSvc) EnableCentralOrdering(centralTenant string) error {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"
	enableCentralOrdering, err := cs.getEnableCentralOrderingByKey(centralTenant, centralOrderingLookupKey)
	if err != nil {
		return err
	}
	if enableCentralOrdering {
		slog.Info(cs.Action.Name, "text", "Central ordering is already enabled", "tenant", centralTenant)
		return nil
	}

	payload, err := json.Marshal(map[string]any{
		"key":   centralOrderingLookupKey,
		"value": "true",
	})
	if err != nil {
		return err
	}

	requestURL := cs.Action.GetRequestURL(constant.KongPort, "/orders-storage/settings")
	headers := helpers.TenantSecureApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)
	err = cs.HTTPClient.PostReturnNoContent(requestURL, payload, headers)
	if err != nil {
		return err
	}
	slog.Info(cs.Action.Name, "text", "Enabled central ordering", "tenant", centralTenant)

	return nil
}

func (cs *ConsortiumSvc) getEnableCentralOrderingByKey(centralTenant string, key string) (bool, error) {
	requestURL := cs.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/orders-storage/settings?query=key==%s&limit=1", key))
	headers := helpers.TenantSecureApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)

	var response models.SettingsResponse
	if err := cs.HTTPClient.GetRetryReturnStruct(requestURL, headers, &response); err != nil {
		return false, err
	}

	if len(response.Settings) == 0 {
		return false, nil
	}

	value := response.Settings[0].Value
	enableCentralOrdering, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return enableCentralOrdering, nil
}
