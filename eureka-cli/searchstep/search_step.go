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

func (ss *SearchStep) ReindexInventoryRecords(tenant string, accessToken string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	for _, record := range []string{"authority", "location", "linked-data-instance", "linked-data-work", "linked-data-hub"} {
		bytes, err := json.Marshal(map[string]any{"recreateIndex": "true", "resourceName": record})
		if err != nil {
			slog.Error(ss.Action.Name, "error", err)
			panic(err)
		}

		reindexJobMap := ss.HTTPClient.DoPostReturnMapStringAny(fmt.Sprintf(ss.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/search/index/inventory/reindex"), false, bytes, headers)
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
}

func (ss *SearchStep) ReindexInstanceRecords(tenant string, accessToken string) {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	bytes, err := json.Marshal(map[string]any{})
	if err != nil {
		slog.Error(ss.Action.Name, "error", err)
		panic(err)
	}

	ss.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ss.HTTPClient.GetGatewayURL(), constant.GatewayPort, "/search/index/instance-records/reindex/full"), false, bytes, headers)

	slog.Info(ss.Action.Name, "text", fmt.Sprintf("Reindexed instance records for %s tenant", tenant))
}
