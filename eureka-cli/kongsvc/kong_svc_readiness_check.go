package kongsvc

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/folio-org/eureka-cli/constant"
)

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
				slog.Info(ks.Action.Name, "text", "Kong route is ready", "id", route.ID, "expression", route.Expression)
			}
			break
		}

		if retryCount == constant.KongRouteReadinessMaxRetries {
			return fmt.Errorf("routes are not ready and out of retries, expected routes: %d, actual: %d", expected, actual)
		}

		slog.Info(ks.Action.Name, "text", "Kong routes are unready", "retryCount", retryCount, "maxRetries", constant.KongRouteReadinessMaxRetries)
		time.Sleep(constant.KongReadinessWait)
	}

	return nil
}
