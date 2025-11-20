package searchsvc

import (
	"encoding/json"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
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
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ss.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}

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
		if err := ss.HTTPClient.PostReturnStruct(requestURL, payload, headers, &job); err != nil {
			slog.Warn(ss.Action.Name, "text", err)
			continue
		}
		if err := ss.validateInventoryRecordsResponse(job); err != nil {
			slog.Warn(ss.Action.Name, "text", "Reindex inventory records was unsuccessful", "tenant", tenantName, "record", record, "error", err)
			continue
		}

		slog.Info(ss.Action.Name, "text", "Reindexed inventory records", "tenant", tenantName, "record", record, "id", job.ID, "status", job.JobStatus)
	}

	return nil
}

func (ss *SearchSvc) validateInventoryRecordsResponse(job models.ReindexJobResponse) error {
	if len(job.Errors) > 0 {
		jobErrors := make([]any, len(job.Errors))
		for i, err := range job.Errors {
			jobErrors[i] = err
		}
		return errors.ReindexJobHasErrors(jobErrors)
	}
	if job.ID == "" {
		return errors.ReindexJobIDBlank()
	}

	return nil
}

func (ss *SearchSvc) ReindexInstanceRecords(tenantName string) error {
	payload, err := json.Marshal(map[string]any{})
	if err != nil {
		return err
	}

	requestURL := ss.Action.GetRequestURL(constant.KongPort, "/search/index/instance-records/reindex/full")
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, ss.Action.KeycloakAccessToken)
	if err != nil {
		return err
	}
	if err := ss.HTTPClient.PostReturnNoContent(requestURL, payload, headers); err != nil {
		return err
	}
	slog.Info(ss.Action.Name, "text", "Reindexed instance records for tenant", "tenant", tenantName)

	return nil
}
