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
		expressions = []string{`(http.path == "/applications" && http.method == "POST")`,
			`(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "POST")`,
			`(http.path == "/entitlements" && http.method == "POST")`,
		}
		expected = len(expressions)
	)
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
