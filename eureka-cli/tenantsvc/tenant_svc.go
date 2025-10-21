package tenantsvc

import (
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/consortiumsvc"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/spf13/viper"
)

type TenantSvc struct {
	Action        *action.Action
	ConsortiumSvc *consortiumsvc.ConsortiumSvc
}

func New(action *action.Action, consortiumSvc *consortiumsvc.ConsortiumSvc) *TenantSvc {
	return &TenantSvc{
		Action:        action,
		ConsortiumSvc: consortiumSvc,
	}
}

func (ts *TenantSvc) GetTenantParameters(consortiumName string, tenants map[string]any) (string, error) {
	if consortiumName == constant.NoneConsortium {
		return "loadReference=true,loadSample=true", nil
	}

	centralTenant := ts.ConsortiumSvc.GetConsortiumCentralTenant(consortiumName, tenants)
	if centralTenant == "" {
		return "", fmt.Errorf("%s consortium does not contain a central tenant", consortiumName)
	}

	return fmt.Sprintf("loadReference=true,loadSample=true,centralTenantId=%s", centralTenant), nil
}

func (ts *TenantSvc) SetDefaultConfigTenantParams(rp *runparams.RunParams, tenant string) error {
	tenants := viper.GetStringMap(field.Tenants)
	if tenants == nil || tenants[tenant] == nil {
		return fmt.Errorf("found not tenant in the config or by %s tenant", tenant)
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

	slog.Info(ts.Action.Name, "text", "Setting default tenant config params", "tenant", tenant, "singleTenant", rp.SingleTenant, "enableECSRequests", rp.EnableECSRequests, "platformCompleteURL", rp.PlatformCompleteURL)

	return nil
}
