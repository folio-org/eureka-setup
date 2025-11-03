package kongsvc

import (
	"fmt"
	"net/http"
	"slices"

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
	CheckRouteExists(routeID string) (bool, *models.KongRoute, error)
	ListAllRoutes() ([]models.KongRoute, error)
	FindRouteByExpressions(expressions []string) ([]*models.KongRoute, error)
}

// KongSvc provides functionality for Kong API gateway operations
type KongSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientGetManager
}

// New creates a new KongSvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientGetManager) KongProcessor {
	return &KongSvc{Action: action, HTTPClient: httpClient}
}

func (ks *KongSvc) CheckRouteExists(routeID string) (bool, *models.KongRoute, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongAdminPort, fmt.Sprintf("/routes/%s", routeID))
	httpResponse, err := ks.HTTPClient.GetReturnResponse(requestURL, nil)
	if err != nil {
		return false, nil, err
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode == http.StatusNotFound {
		return false, nil, nil
	}
	if httpResponse.StatusCode != http.StatusOK {
		return false, nil, errors.KongAdminAPIFailed(httpResponse.StatusCode, httpResponse.Status)
	}

	var route models.KongRoute
	err = ks.HTTPClient.GetReturnStruct(requestURL, nil, &route)
	if err != nil {
		return false, nil, err
	}

	return true, &route, nil
}

func (ks *KongSvc) ListAllRoutes() ([]models.KongRoute, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongAdminPort, "/routes")

	var response models.KongRoutesResponse
	err := ks.HTTPClient.GetReturnStruct(requestURL, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
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
