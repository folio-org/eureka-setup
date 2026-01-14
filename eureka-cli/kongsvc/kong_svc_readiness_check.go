package kongsvc

import (
	"log/slog"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
)

// KongRouteReadinessChecker defines the interface for Kong route readiness check operations
type KongRouteReadinessChecker interface {
	CheckRouteReadiness() error
}

func (ks *KongSvc) CheckRouteReadiness() error {
	var (
		expressions = []string{
			// Applications
			`(http.path == "/applications" && http.method == "GET")`,
			`(http.path == "/applications" && http.method == "POST")`,
			`(http.path ~ "^/applications/([^/]+)$" && http.method == "DELETE")`,

			// Module Discovery
			`(http.path == "/modules/discovery" && http.method == "GET")`,
			`(http.path == "/modules/discovery" && http.method == "POST")`,
			`(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "PUT")`,

			// Tenants
			`(http.path == "/tenants" && http.method == "GET")`,
			`(http.path == "/tenants" && http.method == "POST")`,
			`(http.path ~ "^/tenants/([^/]+)$" && http.method == "DELETE")`,

			// Tenant Entitlement
			`(http.path == "/entitlements" && http.method == "GET")`,
			`(http.path == "/entitlements" && http.method == "POST")`,
			`(http.path == "/entitlements" && http.method == "PUT")`,
			`(http.path == "/entitlements" && http.method == "DELETE")`,
		}
		expected     = len(expressions)
		maxRetries   = helpers.DefaultInt(ks.ReadinessMaxRetries, constant.KongRouteReadinessMaxRetries)
		waitDuration = helpers.DefaultDuration(ks.ReadinessWait, constant.KongReadinessWait)
	)
	slog.Info(ks.Action.Name, "text", "Preparing route readiness check", "expected", expected)
	for retryCount := range maxRetries {
		matchedRoutes, _ := ks.FindRouteByExpressions(expressions)
		if len(matchedRoutes) == expected {
			for _, route := range matchedRoutes {
				slog.Info(ks.Action.Name, "text", "Kong route is ready", "expression", route.Expression)
			}
			return nil
		}

		slog.Warn(ks.Action.Name, "text", "Kong routes are unready", "count", retryCount, "max", maxRetries)
		time.Sleep(waitDuration)
	}

	return errors.KongRoutesNotReady(expected)
}
