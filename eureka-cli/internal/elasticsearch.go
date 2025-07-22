package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

func ReindexInventoryRecords(commandName string, enableDebug bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	for _, record := range []string{"authority", "location", "linked-data-instance", "linked-data-work", "linked-data-hub"} {
		bytes, err := json.Marshal(map[string]any{"recreateIndex": "true", "resourceName": record})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "json.Marshal error")
			panic(err)
		}

		reindexJobMap := DoPostReturnMapStringAny(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/search/index/inventory/reindex"), enableDebug, false, bytes, headers)
		if reindexJobMap["errors"] != nil {
			errorType := reindexJobMap["errors"].([]any)[0].(map[string]any)["type"]
			slog.Warn(commandName, GetFuncName(), fmt.Sprintf("Failed to reindex inventory records for %s tenant and %s record, error type: %s", tenant, record, errorType))
			continue
		}
		if len(reindexJobMap) == 0 {
			slog.Warn(commandName, GetFuncName(), fmt.Sprintf("Failed to reindex inventory records for %s tenant and %s record, no response", tenant, record))
			continue
		}

		jobId := reindexJobMap["id"]
		jobStatus := reindexJobMap["jobStatus"]

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Reindexed inventory records for %s tenant and %s record, job id: %s, job status: %s", tenant, record, jobId, jobStatus))
	}
}

func ReindexInstanceRecords(commandName string, enableDebug bool, tenant string, accessToken string) {
	headers := map[string]string{ContentTypeHeader: JsonContentType, TenantHeader: tenant, TokenHeader: accessToken}

	bytes, err := json.Marshal(map[string]any{})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "json.Marshal error")
		panic(err)
	}

	DoPostReturnNoContent(commandName, fmt.Sprintf(GetGatewayUrlTemplate(commandName), GatewayPort, "/search/index/instance-records/reindex/full"), enableDebug, false, bytes, headers)

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Reindexed instance records for %s tenant", tenant))
}
