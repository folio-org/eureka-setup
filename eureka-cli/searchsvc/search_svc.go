package searchsvc

import (
	"encoding/json"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
)

type SearchSvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *SearchSvc {
	return &SearchSvc{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (ss *SearchSvc) ReindexInventoryRecords(tenant string, accessToken string) error {
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

		reindexJobMap, err := ss.HTTPClient.PostReturnMapStringAny(ss.Action.CreateURL(constant.KongPort, "/search/index/inventory/reindex"), b, headers)
		if err != nil {
			slog.Warn(ss.Action.Name, "text", err)
			continue
		}

		if reindexJobMap["errors"] != nil {
			errorType := reindexJobMap["errors"].([]any)[0].(map[string]any)["type"]
			slog.Warn(ss.Action.Name, "text", "Failed to reindex inventory records with error type", "tenant", tenant, "record", record, "errorType", errorType)
			continue
		}
		if len(reindexJobMap) == 0 {
			slog.Warn(ss.Action.Name, "text", "Failed to reindex inventory records with no response", "tenant", tenant, "record", record)
			continue
		}

		jobId := reindexJobMap["id"]
		jobStatus := reindexJobMap["jobStatus"]

		slog.Info(ss.Action.Name, "text", "Reindexed inventory records", "tenant", tenant, "record", record, "jobId", jobId, "jobStatus", jobStatus)
	}

	return nil
}

func (ss *SearchSvc) ReindexInstanceRecords(tenant string, accessToken string) error {
	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	bytes, err := json.Marshal(map[string]any{})
	if err != nil {
		return err
	}

	err = ss.HTTPClient.PostReturnNoContent(ss.Action.CreateURL(constant.KongPort, "/search/index/instance-records/reindex/full"), bytes, headers)
	if err != nil {
		return err
	}

	slog.Info(ss.Action.Name, "text", "Reindexed instance records for tenant", "tenant", tenant)

	return nil
}
