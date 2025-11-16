package kongsvc

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
)

// KongProcessor defines the interface for Kong service operations
type KongProcessor interface {
	KongRouteReader
	KongRouteReadinessChecker
}

// KongRouteReader defines the interface for Kong route read operations
type KongRouteReader interface {
	ListAllRoutes() ([]models.KongRoute, error)
	FindRouteByExpressions(expressions []string) ([]*models.KongRoute, error)
	CheckRouteExists(routeID string) (bool, *models.KongRoute, error)
}

// KongSvc provides functionality for Kong API gateway operations
type KongSvc struct {
	Action     *action.Action
	HTTPClient interface {
		httpclient.HTTPClientGetManager
		httpclient.HTTPClientPinger
	}
	ReadinessMaxRetries int
	ReadinessWait       time.Duration
}

// New creates a new KongSvc instance
func New(action *action.Action, httpClient interface {
	httpclient.HTTPClientGetManager
	httpclient.HTTPClientPinger
}) KongProcessor {
	return &KongSvc{Action: action, HTTPClient: httpClient}
}

func (ks *KongSvc) ListAllRoutes() ([]models.KongRoute, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongAdminPort, "/routes")

	var decodedResponse models.KongRoutesResponse
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, nil, &decodedResponse); err != nil {
		return nil, err
	}

	return decodedResponse.Data, nil
}

func (ks *KongSvc) FindRouteByExpressions(expressions []string) ([]*models.KongRoute, error) {
	allRoutes, err := ks.ListAllRoutes()
	if err != nil {
		return nil, err
	}

	var routes []*models.KongRoute
	for _, route := range allRoutes {
		if slices.Contains(expressions, route.Expression) {
			routes = append(routes, &route)
		}
	}

	return routes, nil
}

func (ks *KongSvc) CheckRouteExists(routeID string) (bool, *models.KongRoute, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongAdminPort, fmt.Sprintf("/routes/%s", routeID))
	statusCode, err := ks.HTTPClient.Ping(requestURL)
	if err != nil {
		return false, nil, err
	}

	if statusCode == http.StatusNotFound {
		return false, nil, nil
	}
	if statusCode != http.StatusOK {
		return false, nil, errors.KongAdminAPIFailed(statusCode, http.StatusText(statusCode))
	}

	var decodedResponse models.KongRoute
	if err := ks.HTTPClient.GetRetryReturnStruct(requestURL, nil, &decodedResponse); err != nil {
		return false, nil, err
	}

	return true, &decodedResponse, nil
}
