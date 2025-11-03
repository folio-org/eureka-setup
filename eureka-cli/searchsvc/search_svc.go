package searchsvc

import (
	"encoding/json"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
)

// SearchProcessor defines the interface for search and reindexing operations
type SearchProcessor interface {
	ReindexInventoryRecords(tenantName string) error
	ReindexInstanceRecords(tenantName string) error
}

// SearchSvc provides functionality for managing search indices and reindexing operations
type SearchSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
}

// New creates a new SearchSvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner) *SearchSvc {
	return &SearchSvc{Action: action, HTTPClient: httpClient}
}

func (ss *SearchSvc) ReindexInventoryRecords(tenantName string) error {
	requestURL := ss.Action.GetRequestURL(constant.KongPort, "/search/index/inventory/reindex")
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ss.Action.KeycloakAccessToken)
	inventoryRecords := []string{"authority", "location", "linked-data-instance", "linked-data-work", "linked-data-hub"}
	for _, record := range inventoryRecords {
		payload, err := json.Marshal(map[string]any{
			"recreateIndex": "true",
			"resourceName":  record,
		})
		if err != nil {
			return err
		}

		var job models.ReindexJobResponse
		err = ss.HTTPClient.PostReturnStruct(requestURL, payload, headers, &job)
		if err != nil {
			slog.Warn(ss.Action.Name, "text", err)
			continue
		}
		if len(job.Errors) > 0 {
			slog.Warn(ss.Action.Name, "text", "Failed to reindex inventory records with error type", "tenant", tenantName, "record", record, "errorType", job.Errors[0].Type)
			continue
		}
		if job.ID == "" {
			slog.Warn(ss.Action.Name, "text", "Failed to reindex inventory records with no job ID", "tenant", tenantName, "record", record)
			continue
		}
		slog.Info(ss.Action.Name, "text", "Reindexed inventory records", "tenant", tenantName, "record", record, "jobId", job.ID, "jobStatus", job.JobStatus)
	}

	return nil
}

func (ss *SearchSvc) ReindexInstanceRecords(tenantName string) error {
	payload, err := json.Marshal(map[string]any{})
	if err != nil {
		return err
	}

	requestURL := ss.Action.GetRequestURL(constant.KongPort, "/search/index/instance-records/reindex/full")
	headers := helpers.TenantSecureApplicationJSONHeaders(tenantName, ss.Action.KeycloakAccessToken)
	err = ss.HTTPClient.PostReturnNoContent(requestURL, payload, headers)
	if err != nil {
		return err
	}
	slog.Info(ss.Action.Name, "text", "Reindexed instance records for tenant", "tenant", tenantName)

	return nil
}
