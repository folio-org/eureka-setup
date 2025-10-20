package searchstep

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
)

type SearchStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *SearchStep {
	return &SearchStep{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (ss *SearchStep) ReindexInventoryRecords(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	records := []string{"authority", "location", "linked-data-instance", "linked-data-work", "linked-data-hub"}

	for _, record := range records {
		b, err := json.Marshal(map[string]any{"recreateIndex": "true", "resourceName": record})
		if err != nil {
			return err
		}

		reindexJobMap, err := ss.HTTPClient.DoPostReturnMapStringAny(fmt.Sprintf(ss.Action.GatewayURL, constant.KongPort, "/search/index/inventory/reindex"), b, headers)
		if err != nil {
			return err
		}

		if reindexJobMap["errors"] != nil {
			errorType := reindexJobMap["errors"].([]any)[0].(map[string]any)["type"]
			slog.Warn(ss.Action.Name, "text", fmt.Sprintf("failed to reindex inventory records for %s tenant and %s record, error type: %s", tenant, record, errorType))
			continue
		}
		if len(reindexJobMap) == 0 {
			slog.Warn(ss.Action.Name, "text", fmt.Sprintf("failed to reindex inventory records for %s tenant and %s record, no response", tenant, record))
			continue
		}

		jobId := reindexJobMap["id"]
		jobStatus := reindexJobMap["jobStatus"]

		slog.Info(ss.Action.Name, "text", fmt.Sprintf("reindexed inventory records for %s tenant and %s record, job id: %s, job status: %s", tenant, record, jobId, jobStatus))
	}

	return nil
}

func (ss *SearchStep) ReindexInstanceRecords(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	bytes, err := json.Marshal(map[string]any{})
	if err != nil {
		return err
	}

	ss.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ss.Action.GatewayURL, constant.KongPort, "/search/index/instance-records/reindex/full"), bytes, headers)

	slog.Info(ss.Action.Name, "text", fmt.Sprintf("Reindexed instance records for %s tenant", tenant))

	return nil
}
