package tenantstep

import (
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/consortiumstep"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/spf13/viper"
)

type TenantStep struct {
	Action         *action.Action
	ConsortiumStep *consortiumstep.ConsortiumStep
}

func New(action *action.Action, consortiumStep *consortiumstep.ConsortiumStep) *TenantStep {
	return &TenantStep{
		Action:         action,
		ConsortiumStep: consortiumStep,
	}
}

func (ts *TenantStep) GetTenantParameters(consortiumName string, tenants map[string]any) string {
	if consortiumName == constant.NoneConsortium {
		return "loadReference=true,loadSample=true"
	}

	centralTenant := ts.ConsortiumStep.GetConsortiumCentralTenant(consortiumName, tenants)
	if centralTenant == "" {
		helpers.LogErrorPanic(ts.Action, fmt.Errorf("%s consortium does not contain a central tenant", consortiumName))
		return ""
	}

	return fmt.Sprintf("loadReference=true,loadSample=true,centralTenantId=%s", centralTenant)
}

func (ts *TenantStep) SetDefaultConfigTenantParams(rp *runparams.RunParams, tenant string) {
	tenants := viper.GetStringMap(field.Tenants)
	if tenants == nil || tenants[tenant] == nil {
		helpers.LogDebug(ts.Action, fmt.Errorf("found not tenant in the config or by %s tenant", tenant))
		return
	}

	var tenantMap = tenants[tenant].(map[string]any)
	if tenantMap[field.TenantsSingleTenantEntry] != nil {
		rp.SingleTenant = tenantMap[field.TenantsSingleTenantEntry].(bool)
	}
	if tenantMap[field.TenantsEnableEcsRequestEntry] != nil {
		rp.EnableECSRequests = tenantMap[field.TenantsEnableEcsRequestEntry].(bool)
	}
	if tenantMap[field.TenantsPlatformCompleteURLEntry] != nil {
		rp.PlatformCompleteURL = tenantMap[field.TenantsPlatformCompleteURLEntry].(string)
	}

	preparedParams := []any{tenant, rp.SingleTenant, rp.EnableECSRequests, rp.PlatformCompleteURL}
	slog.Info(ts.Action.Name, "text", fmt.Sprintf("Setting default tenant config params, tenant: %s, singleTenant: %t, enableECSRequests: %t, platformCompleteURL: %s", preparedParams...))
}
