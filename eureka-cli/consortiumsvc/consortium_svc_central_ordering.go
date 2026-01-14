package consortiumsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
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
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}
	if err := cs.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(cs.Action.Name, "text", "Enabled central ordering", "tenant", centralTenant)

	return nil
}

func (cs *ConsortiumSvc) getEnableCentralOrderingByKey(centralTenant string, key string) (bool, error) {
	requestURL := cs.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/orders-storage/settings?query=key==%s&limit=1", key))
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(centralTenant, cs.Action.KeycloakAccessToken)
	if err != nil {
		return false, err
	}

	var decodedResponse models.SettingsResponse
	if err := cs.HTTPClient.GetRetryReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return false, err
	}
	if len(decodedResponse.Settings) == 0 {
		return false, nil
	}

	enableCentralOrdering, err := strconv.ParseBool(decodedResponse.Settings[0].Value)
	if err != nil {
		return false, err
	}

	return enableCentralOrdering, nil
}
