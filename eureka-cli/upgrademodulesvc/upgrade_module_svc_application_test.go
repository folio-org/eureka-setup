package upgrademodulesvc

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBackendModules_DiscoveryUsesConfiguredPrivatePort(t *testing.T) {
	// Arrange
	mockAction := testhelpers.NewMockAction()
	mockAction.ConfigBackendModules = map[string]any{
		"mod-agreements": map[string]any{"private-port": 8080},
	}
	svc := &UpgradeModuleSvc{Action: mockAction}
	modules := []any{
		map[string]any{"id": "mod-agreements-7.4.0", "name": "mod-agreements", "version": "7.4.0"},
	}

	// Act
	_, discovery, oldModuleID, err := svc.UpdateBackendModules("mod-agreements", "7.4.1", true, modules)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "mod-agreements-7.4.0", oldModuleID)
	assert.Equal(t, "http://mod-agreements-sc.eureka:8080", discovery[0]["location"])
}

func TestUpdateBackendModules_DiscoveryUsesPortServerAlias(t *testing.T) {
	// Arrange
	mockAction := testhelpers.NewMockAction()
	mockAction.ConfigBackendModules = map[string]any{
		"mod-agreements": map[string]any{"private-port": 8080, "port-server": 9090},
	}
	svc := &UpgradeModuleSvc{Action: mockAction}
	modules := []any{
		map[string]any{"id": "mod-agreements-7.4.0", "name": "mod-agreements", "version": "7.4.0"},
	}

	// Act
	_, discovery, _, err := svc.UpdateBackendModules("mod-agreements", "7.4.1", true, modules)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "http://mod-agreements-sc.eureka:9090", discovery[0]["location"])
}

func TestUpdateBackendModules_DiscoveryDefaultsToPrivateServerPort(t *testing.T) {
	// Arrange
	svc := &UpgradeModuleSvc{Action: testhelpers.NewMockAction()}
	modules := []any{
		map[string]any{"id": "mod-orders-13.0.0", "name": "mod-orders", "version": "13.0.0"},
	}

	// Act
	_, discovery, _, err := svc.UpdateBackendModules("mod-orders", "13.0.1", true, modules)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "http://mod-orders-sc.eureka:8081", discovery[0]["location"])
}
