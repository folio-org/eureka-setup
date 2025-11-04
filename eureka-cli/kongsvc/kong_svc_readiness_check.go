package kongsvc

import (
	"log/slog"
	"time"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
)

// KongRouteReadinessChecker defines the interface for Kong route readiness check operations
type KongRouteReadinessChecker interface {
	CheckRouteReadiness() error
}

func (ks *KongSvc) CheckRouteReadiness() error {
	var (
		// TODO Add more expressions
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
		expected = len(expressions)
	)
	slog.Info(ks.Action.Name, "text", "Preparing route readiness check", "expected", expected)
	for retryCount := range constant.KongRouteReadinessMaxRetries {
		matchedRoutes, _ := ks.FindRouteByExpressions(expressions)
		actual := len(matchedRoutes)
		if actual == expected {
			for _, route := range matchedRoutes {
				slog.Info(ks.Action.Name, "text", "Kong route is ready", "expression", route.Expression)
			}
			break
		}

		if retryCount == constant.KongRouteReadinessMaxRetries {
			return errors.KongRoutesNotReady(actual, expected)
		}

		slog.Info(ks.Action.Name, "text", "Kong routes are unready", "retryCount", retryCount, "maxRetries", constant.KongRouteReadinessMaxRetries)
		time.Sleep(constant.KongReadinessWait)
	}

	return nil
}
