package tenantsvc

import (
	"fmt"
	"log/slog"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/consortiumsvc"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
)

// TenantProcessor defines the interface for tenant-related operations
type TenantProcessor interface {
	GetEntitlementTenantParameters(consortiumName string) (string, error)
	SetConfigTenantParams(tenantName string) error
}

// TenantSvc provides functionality for managing tenant configurations and parameters
type TenantSvc struct {
	Action        *action.Action
	ConsortiumSvc consortiumsvc.ConsortiumProcessor
}

// New creates a new TenantSvc instance
func New(action *action.Action, consortiumSvc consortiumsvc.ConsortiumProcessor) *TenantSvc {
	return &TenantSvc{Action: action, ConsortiumSvc: consortiumSvc}
}

func (ts *TenantSvc) GetEntitlementTenantParameters(consortiumName string) (string, error) {
	if consortiumName == constant.NoneConsortium {
		return "loadReference=true,loadSample=true", nil
	}

	centralTenant := ts.ConsortiumSvc.GetConsortiumCentralTenant(consortiumName)
	if centralTenant == "" {
		return "", errors.ConsortiumMissingCentralTenant(consortiumName)
	}

	return fmt.Sprintf("loadReference=true,loadSample=true,centralTenantId=%s", centralTenant), nil
}

func (ts *TenantSvc) SetConfigTenantParams(tenantName string) error {
	if ts.Action.ConfigTenants == nil || ts.Action.ConfigTenants[tenantName] == nil {
		return errors.TenantNotFound(tenantName)
	}

	var configTenant = ts.Action.ConfigTenants[tenantName].(map[string]any)
	helpers.SetBool(configTenant, field.TenantsSingleTenantEntry, &ts.Action.Param.SingleTenant)
	helpers.SetBool(configTenant, field.TenantsEnableEcsRequestEntry, &ts.Action.Param.EnableECSRequests)
	helpers.SetString(configTenant, field.TenantsPlatformCompleteURLEntry, &ts.Action.Param.PlatformCompleteURL)
	slog.Info(ts.Action.Name, "text", "Setting default tenant config params", "tenant", tenantName)

	return nil
}
