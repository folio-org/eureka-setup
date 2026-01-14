package kongsvc_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/j011195/eureka-setup/eureka-cli/kongsvc"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := &testhelpers.MockHTTPClient{}

	// Act
	svc := kongsvc.New(action, mockHTTP)

	// Assert
	assert.NotNil(t, svc)
}

func TestCheckRouteExists_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routeID := "route-123"
	expectedRoute := &models.KongRoute{
		ID:         routeID,
		Name:       "test-route",
		Expression: `(http.path == "/test")`,
	}

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		})).
		Return(http.StatusOK, nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoute)
			*target = *expectedRoute
		}).
		Return(nil)

	// Act
	exists, route, err := svc.CheckRouteExists(routeID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, route)
	assert.Equal(t, routeID, route.ID)
	assert.Equal(t, "test-route", route.Name)
	mockHTTP.AssertExpectations(t)
}

func TestCheckRouteExists_NotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routeID := "nonexistent-route"

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		})).
		Return(http.StatusNotFound, nil)

	// Act
	exists, route, err := svc.CheckRouteExists(routeID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, route)
	mockHTTP.AssertExpectations(t)
}

func TestCheckRouteExists_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routeID := "route-123"
	expectedError := errors.New("HTTP request failed")

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		})).
		Return(0, expectedError)

	// Act
	exists, route, err := svc.CheckRouteExists(routeID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
	assert.Nil(t, route)
	mockHTTP.AssertExpectations(t)
}

func TestCheckRouteExists_InternalServerError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routeID := "route-123"

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		})).
		Return(http.StatusInternalServerError, nil)

	// Act
	exists, route, err := svc.CheckRouteExists(routeID)

	// Assert
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Nil(t, route)
	assert.Contains(t, err.Error(), "kong admin API failed")
	mockHTTP.AssertExpectations(t)
}

func TestCheckRouteExists_GetStructError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routeID := "route-123"
	expectedError := errors.New("Failed to decode response")

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		})).
		Return(http.StatusOK, nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes/"+routeID)
		}),
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	exists, route, err := svc.CheckRouteExists(routeID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
	assert.Nil(t, route)
	mockHTTP.AssertExpectations(t)
}

func TestListAllRoutes_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{
		Data: []models.KongRoute{
			{
				ID:         "route-1",
				Name:       "applications-get",
				Expression: `(http.path == "/applications" && http.method == "GET")`,
			},
			{
				ID:         "route-2",
				Name:       "tenants-post",
				Expression: `(http.path == "/tenants" && http.method == "POST")`,
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/routes")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	// Act
	routes, err := svc.ListAllRoutes()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, routes, 2)
	assert.Equal(t, "route-1", routes[0].ID)
	assert.Equal(t, "route-2", routes[1].ID)
	mockHTTP.AssertExpectations(t)
}

func TestListAllRoutes_EmptyResponse(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{Data: []models.KongRoute{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	// Act
	routes, err := svc.ListAllRoutes()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, routes, 0)
	mockHTTP.AssertExpectations(t)
}

func TestListAllRoutes_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	expectedError := errors.New("HTTP request failed")

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	routes, err := svc.ListAllRoutes()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, routes)
	mockHTTP.AssertExpectations(t)
}

func TestFindRouteByExpressions_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{
		Data: []models.KongRoute{
			{
				ID:         "route-1",
				Name:       "applications-get",
				Expression: `(http.path == "/applications" && http.method == "GET")`,
			},
			{
				ID:         "route-2",
				Name:       "tenants-post",
				Expression: `(http.path == "/tenants" && http.method == "POST")`,
			},
			{
				ID:         "route-3",
				Name:       "other-route",
				Expression: `(http.path == "/other")`,
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	expressions := []string{
		`(http.path == "/applications" && http.method == "GET")`,
		`(http.path == "/tenants" && http.method == "POST")`,
	}

	// Act
	routes, err := svc.FindRouteByExpressions(expressions)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, routes, 2)
	assert.Equal(t, "route-1", routes[0].ID)
	assert.Equal(t, "route-2", routes[1].ID)
	mockHTTP.AssertExpectations(t)
}

func TestFindRouteByExpressions_NoMatches(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{
		Data: []models.KongRoute{
			{
				ID:         "route-1",
				Name:       "applications-get",
				Expression: `(http.path == "/applications" && http.method == "GET")`,
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	expressions := []string{
		`(http.path == "/nonexistent")`,
	}

	// Act
	routes, err := svc.FindRouteByExpressions(expressions)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, routes, 0)
	mockHTTP.AssertExpectations(t)
}

func TestFindRouteByExpressions_ListAllRoutesError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	expectedError := errors.New("Failed to list routes")

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	expressions := []string{
		`(http.path == "/applications")`,
	}

	// Act
	routes, err := svc.FindRouteByExpressions(expressions)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, routes)
	mockHTTP.AssertExpectations(t)
}

func TestCheckRouteReadiness_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{
		Data: []models.KongRoute{
			{ID: "route-1", Expression: `(http.path == "/applications" && http.method == "GET")`},
			{ID: "route-2", Expression: `(http.path == "/applications" && http.method == "POST")`},
			{ID: "route-3", Expression: `(http.path ~ "^/applications/([^/]+)$" && http.method == "DELETE")`},
			{ID: "route-4", Expression: `(http.path == "/modules/discovery" && http.method == "GET")`},
			{ID: "route-5", Expression: `(http.path == "/modules/discovery" && http.method == "POST")`},
			{ID: "route-6", Expression: `(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "PUT")`},
			{ID: "route-7", Expression: `(http.path == "/tenants" && http.method == "GET")`},
			{ID: "route-8", Expression: `(http.path == "/tenants" && http.method == "POST")`},
			{ID: "route-9", Expression: `(http.path ~ "^/tenants/([^/]+)$" && http.method == "DELETE")`},
			{ID: "route-10", Expression: `(http.path == "/entitlements" && http.method == "GET")`},
			{ID: "route-11", Expression: `(http.path == "/entitlements" && http.method == "POST")`},
			{ID: "route-12", Expression: `(http.path == "/entitlements" && http.method == "PUT")`},
			{ID: "route-13", Expression: `(http.path == "/entitlements" && http.method == "DELETE")`},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	// Act
	err := svc.CheckRouteReadiness()

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestFindRouteByExpressions_AllExpressionsMatched(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := kongsvc.New(action, mockHTTP)

	routesResponse := models.KongRoutesResponse{
		Data: []models.KongRoute{
			{ID: "route-1", Expression: `(http.path == "/applications" && http.method == "GET")`},
			{ID: "route-2", Expression: `(http.path == "/applications" && http.method == "POST")`},
			{ID: "route-3", Expression: `(http.path ~ "^/applications/([^/]+)$" && http.method == "DELETE")`},
			{ID: "route-4", Expression: `(http.path == "/modules/discovery" && http.method == "GET")`},
			{ID: "route-5", Expression: `(http.path == "/modules/discovery" && http.method == "POST")`},
			{ID: "route-6", Expression: `(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "PUT")`},
			{ID: "route-7", Expression: `(http.path == "/tenants" && http.method == "GET")`},
			{ID: "route-8", Expression: `(http.path == "/tenants" && http.method == "POST")`},
			{ID: "route-9", Expression: `(http.path ~ "^/tenants/([^/]+)$" && http.method == "DELETE")`},
			{ID: "route-10", Expression: `(http.path == "/entitlements" && http.method == "GET")`},
			{ID: "route-11", Expression: `(http.path == "/entitlements" && http.method == "POST")`},
			{ID: "route-12", Expression: `(http.path == "/entitlements" && http.method == "PUT")`},
			{ID: "route-13", Expression: `(http.path == "/entitlements" && http.method == "DELETE")`},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KongRoutesResponse)
			*target = routesResponse
		}).
		Return(nil)

	expressions := []string{
		`(http.path == "/applications" && http.method == "GET")`,
		`(http.path == "/applications" && http.method == "POST")`,
		`(http.path ~ "^/applications/([^/]+)$" && http.method == "DELETE")`,
		`(http.path == "/modules/discovery" && http.method == "GET")`,
		`(http.path == "/modules/discovery" && http.method == "POST")`,
		`(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "PUT")`,
		`(http.path == "/tenants" && http.method == "GET")`,
		`(http.path == "/tenants" && http.method == "POST")`,
		`(http.path ~ "^/tenants/([^/]+)$" && http.method == "DELETE")`,
		`(http.path == "/entitlements" && http.method == "GET")`,
		`(http.path == "/entitlements" && http.method == "POST")`,
		`(http.path == "/entitlements" && http.method == "PUT")`,
		`(http.path == "/entitlements" && http.method == "DELETE")`,
	}

	// Act
	routes, err := svc.FindRouteByExpressions(expressions)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, routes, 13)
	mockHTTP.AssertExpectations(t)
}
