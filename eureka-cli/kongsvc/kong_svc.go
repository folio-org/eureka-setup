package kongsvc

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
)

type KongSvcProcessor interface {
	CheckRouteExists(routeID string) (bool, *models.KongRoute, error)
	ListAllRoutes() ([]models.KongRoute, error)
	FindRouteByExpressions(expressions []string) ([]*models.KongRoute, error)
}

type KongSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientGetManager
}

func New(action *action.Action, httpClient httpclient.HTTPClientGetManager) KongSvcProcessor {
	return &KongSvc{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (ks *KongSvc) CheckRouteExists(routeID string) (bool, *models.KongRoute, error) {
	requestURL := ks.Action.GetRequestURL(constant.KongAdminPort, fmt.Sprintf("/routes/%s", routeID))
	resp, err := ks.HTTPClient.GetReturnResponse(requestURL, nil)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, nil, fmt.Errorf("kong admin API error: %d %s", resp.StatusCode, resp.Status)
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
