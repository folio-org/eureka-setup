package tenantsvc_test

import (
	stderrors "errors"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/tenantsvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock ConsortiumSvc ====================

type MockConsortiumSvc struct {
	mock.Mock
}

// ConsortiumManager interface methods
func (m *MockConsortiumSvc) GetConsortiumByName(centralTenant string, consortiumName string) (any, error) {
	args := m.Called(centralTenant, consortiumName)
	return args.Get(0), args.Error(1)
}

func (m *MockConsortiumSvc) GetConsortiumCentralTenant(consortiumName string) string {
	args := m.Called(consortiumName)
	return args.String(0)
}

func (m *MockConsortiumSvc) GetConsortiumUsers(consortiumName string) map[string]any {
	args := m.Called(consortiumName)
	return args.Get(0).(map[string]any)
}

func (m *MockConsortiumSvc) GetAdminUsername(centralTenant string, consortiumUsers map[string]any) string {
	args := m.Called(centralTenant, consortiumUsers)
	return args.String(0)
}

func (m *MockConsortiumSvc) CreateConsortium(centralTenant string, consortiumName string) (string, error) {
	args := m.Called(centralTenant, consortiumName)
	return args.String(0), args.Error(1)
}

// ConsortiumTenantHandler interface methods
func (m *MockConsortiumSvc) GetSortedConsortiumTenants(consortiumName string) models.SortedConsortiumTenants {
	args := m.Called(consortiumName)
	return args.Get(0).(models.SortedConsortiumTenants)
}

func (m *MockConsortiumSvc) CreateConsortiumTenants(centralTenant string, consortiumID string, consortiumTenants models.SortedConsortiumTenants, adminUsername string) error {
	args := m.Called(centralTenant, consortiumID, consortiumTenants, adminUsername)
	return args.Error(0)
}

// ConsortiumCentralOrderingManager interface methods
func (m *MockConsortiumSvc) EnableCentralOrdering(centralTenant string) error {
	args := m.Called(centralTenant)
	return args.Error(0)
}

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mockConsortiumSvc := new(MockConsortiumSvc)

		// Act
		result := tenantsvc.New(act, mockConsortiumSvc)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, act, result.Action)
		assert.Equal(t, mockConsortiumSvc, result.ConsortiumSvc)
	})
}

// ==================== GetEntitlementTenantParameters Tests ====================

func TestGetEntitlementTenantParameters_NoneConsortium(t *testing.T) {
	t.Run("TestGetEntitlementTenantParameters_NoneConsortium_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		result, err := svc.GetEntitlementTenantParameters(constant.NoneConsortium)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "loadReference=true,loadSample=true", result)
	})
}

func TestGetEntitlementTenantParameters_WithCentralTenant(t *testing.T) {
	t.Run("TestGetEntitlementTenantParameters_WithCentralTenant_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)
		consortiumName := "test-consortium"
		centralTenant := "central-tenant"

		mockConsortiumSvc.On("GetConsortiumCentralTenant", consortiumName).Return(centralTenant)

		// Act
		result, err := svc.GetEntitlementTenantParameters(consortiumName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "loadReference=true,loadSample=true,centralTenantId=central-tenant", result)
		mockConsortiumSvc.AssertExpectations(t)
	})
}

func TestGetEntitlementTenantParameters_NoCentralTenant(t *testing.T) {
	t.Run("TestGetEntitlementTenantParameters_NoCentralTenant_Error_MissingCentralTenant", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)
		consortiumName := "test-consortium"

		mockConsortiumSvc.On("GetConsortiumCentralTenant", consortiumName).Return("")

		// Act
		result, err := svc.GetEntitlementTenantParameters(consortiumName)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "consortium test-consortium does not contain a central tenant")
		mockConsortiumSvc.AssertExpectations(t)
	})
}

// ==================== SetConfigTenantParams Tests ====================

func TestSetConfigTenantParams_Success(t *testing.T) {
	t.Run("TestSetConfigTenantParams_Success_AllFieldsSet", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:  "test-action",
			Param: &action.Param{},
			ConfigTenants: map[string]any{
				"diku": map[string]any{
					"single-tenant":         true,
					"enable-ecs-request":    true,
					"platform-complete-url": "http://localhost:8080",
				},
			},
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("diku")

		// Assert
		assert.NoError(t, err)
		assert.True(t, svc.Action.Param.SingleTenant)
		assert.True(t, svc.Action.Param.EnableECSRequests)
		assert.Equal(t, "http://localhost:8080", svc.Action.Param.PlatformCompleteURL)
	})

	t.Run("TestSetConfigTenantParams_Success_PartialFieldsSet", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:  "test-action",
			Param: &action.Param{},
			ConfigTenants: map[string]any{
				"diku": map[string]any{
					"single-tenant": false,
				},
			},
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("diku")

		// Assert
		assert.NoError(t, err)
		assert.False(t, svc.Action.Param.SingleTenant)
	})

	t.Run("TestSetConfigTenantParams_Success_EmptyConfigTenant", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:  "test-action",
			Param: &action.Param{},
			ConfigTenants: map[string]any{
				"diku": map[string]any{},
			},
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("diku")

		// Assert
		assert.NoError(t, err)
	})
}

func TestSetConfigTenantParams_TenantNotFound(t *testing.T) {
	t.Run("TestSetConfigTenantParams_TenantNotFound_NilConfigTenants", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:          "test-action",
			ConfigTenants: nil,
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("nonexistent")

		// Assert
		assert.Error(t, err)
		assert.True(t, stderrors.Is(err, errors.ErrNotFound))
		assert.Contains(t, err.Error(), "tenant nonexistent in config")
	})

	t.Run("TestSetConfigTenantParams_TenantNotFound_TenantNotInConfig", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigTenants: map[string]any{
				"diku": map[string]any{},
			},
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("nonexistent")

		// Assert
		assert.Error(t, err)
		assert.True(t, stderrors.Is(err, errors.ErrNotFound))
		assert.Contains(t, err.Error(), "tenant nonexistent in config")
	})

	t.Run("TestSetConfigTenantParams_TenantNotFound_TenantIsNil", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigTenants: map[string]any{
				"diku": nil,
			},
		}
		mockConsortiumSvc := new(MockConsortiumSvc)
		svc := tenantsvc.New(act, mockConsortiumSvc)

		// Act
		err := svc.SetConfigTenantParams("diku")

		// Assert
		assert.Error(t, err)
		assert.True(t, stderrors.Is(err, errors.ErrNotFound))
		assert.Contains(t, err.Error(), "tenant diku in config")
	})
}
