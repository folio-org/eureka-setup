package consortiumstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/folio-org/eureka-cli/constant"
)

func (cs *ConsortiumStep) EnableCentralOrdering(centralTenant string, accessToken string) error {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"

	enableCentralOrdering, err := cs.getEnableCentralOrderingByKey(centralTenant, accessToken, centralOrderingLookupKey)
	if err != nil {
		return err
	}

	if enableCentralOrdering {
		slog.Info(cs.Action.Name, "text", fmt.Sprintf("central ordering for %s tenant is already enabled", centralTenant))
		return nil
	}

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	b, err := json.Marshal(map[string]any{"key": centralOrderingLookupKey, "value": "true"})
	if err != nil {
		return err
	}

	cs.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(cs.Action.GatewayURL, constant.KongPort, "/orders-storage/settings"), b, headers)

	slog.Info(cs.Action.Name, "text", fmt.Sprintf("Enabled central ordering for %s tenant", centralTenant))

	return nil
}

func (cs *ConsortiumStep) getEnableCentralOrderingByKey(centralTenant string, accessToken string, key string) (bool, error) {
	requestURL := fmt.Sprintf(cs.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/orders-storage/settings?query=key==%s", key))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	foundSettingsMap, err := cs.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return false, err
	}

	if foundSettingsMap["settings"] == nil || len(foundSettingsMap["settings"].([]any)) == 0 {
		return false, nil
	}

	settings := foundSettingsMap["settings"].([]any)[0]
	value := settings.(map[string]any)["value"].(string)

	enableCentralOrdering, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return enableCentralOrdering, nil
}
