package consortiumsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/folio-org/eureka-cli/constant"
)

func (cs *ConsortiumSvc) EnableCentralOrdering(centralTenant string, accessToken string) error {
	centralOrderingLookupKey := "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS"

	enableCentralOrdering, err := cs.getEnableCentralOrderingByKey(centralTenant, accessToken, centralOrderingLookupKey)
	if err != nil {
		return err
	}

	if enableCentralOrdering {
		slog.Info(cs.Action.Name, "text", "Central ordering for tenant is already enabled", "tenant", centralTenant)
		return nil
	}

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	bb, err := json.Marshal(map[string]any{
		"key":   centralOrderingLookupKey,
		"value": "true",
	})
	if err != nil {
		return err
	}

	err = cs.HTTPClient.PostReturnNoContent(cs.Action.CreateURL(constant.KongPort, "/orders-storage/settings"), bb, headers)
	if err != nil {
		return err
	}

	slog.Info(cs.Action.Name, "text", "Enabled central ordering for tenant", "tenant", centralTenant)

	return nil
}

func (cs *ConsortiumSvc) getEnableCentralOrderingByKey(centralTenant string, accessToken string, key string) (bool, error) {
	requestURL := cs.Action.CreateURL(constant.KongPort, fmt.Sprintf("/orders-storage/settings?query=key==%s", key))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: centralTenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	ss1, err := cs.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return false, err
	}

	if ss1["settings"] == nil || len(ss1["settings"].([]any)) == 0 {
		return false, nil
	}

	ss2 := ss1["settings"].([]any)[0]
	value := ss2.(map[string]any)["value"].(string)

	enableCentralOrdering, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return enableCentralOrdering, nil
}
