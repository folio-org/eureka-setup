package moduleenv_test

import (
	"strings"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/moduleenv"
	"github.com/stretchr/testify/assert"
)

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}

		// Act
		result := moduleenv.New(act)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, act, result.Action)
	})
}

// ==================== VaultEnv Tests ====================

func TestVaultEnv(t *testing.T) {
	t.Run("TestVaultEnv_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{"EXISTING_VAR=value"}
		vaultRootToken := "test-root-token"

		// Act
		result := mv.VaultEnv(env, vaultRootToken)

		// Assert
		assert.Len(t, result, 4)
		assert.Contains(t, result, "EXISTING_VAR=value")
		assert.Contains(t, result, "SECRET_STORE_TYPE=VAULT")
		assert.Contains(t, result, "SECRET_STORE_VAULT_TOKEN=test-root-token")
		assert.Contains(t, result, "SECRET_STORE_VAULT_ADDRESS=http://vault.eureka:8200")
	})

	t.Run("TestVaultEnv_EmptyEnv", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		vaultRootToken := "token123"

		// Act
		result := mv.VaultEnv(env, vaultRootToken)

		// Assert
		assert.Len(t, result, 3)
		assert.Contains(t, result, "SECRET_STORE_TYPE=VAULT")
		assert.Contains(t, result, "SECRET_STORE_VAULT_TOKEN=token123")
	})
}

// ==================== OkapiEnv Tests ====================

func TestOkapiEnv(t *testing.T) {
	t.Run("TestOkapiEnv_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{"EXISTING=value"}
		sidecarName := "mod-inventory-sidecar"
		privatePort := 8081

		// Act
		result := mv.OkapiEnv(env, sidecarName, privatePort)

		// Assert
		assert.Len(t, result, 6)
		assert.Contains(t, result, "EXISTING=value")
		assert.Contains(t, result, "OKAPI_HOST=mod-inventory-sidecar.eureka")
		assert.Contains(t, result, "OKAPI_PORT=8081")
		assert.Contains(t, result, "OKAPI_SERVICE_HOST=mod-inventory-sidecar.eureka")
		assert.Contains(t, result, "OKAPI_SERVICE_URL=http://mod-inventory-sidecar.eureka:8081")
		assert.Contains(t, result, "OKAPI_URL=http://mod-inventory-sidecar.eureka:8081")
	})

	t.Run("TestOkapiEnv_DifferentPort", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		sidecarName := "test-sidecar"
		privatePort := 9000

		// Act
		result := mv.OkapiEnv(env, sidecarName, privatePort)

		// Assert
		assert.Len(t, result, 5)
		assert.Contains(t, result, "OKAPI_PORT=9000")
		assert.Contains(t, result, "OKAPI_URL=http://test-sidecar.eureka:9000")
	})
}

// ==================== DisabledSystemUserEnv Tests ====================

func TestDisabledSystemUserEnv(t *testing.T) {
	t.Run("TestDisabledSystemUserEnv_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{"EXISTING=value"}
		moduleName := "mod-inventory"

		// Act
		result := mv.DisabledSystemUserEnv(env, moduleName)

		// Assert
		assert.Len(t, result, 6)
		assert.Contains(t, result, "EXISTING=value")
		assert.Contains(t, result, "FOLIO_SYSTEM_USER_ENABLED=false")
		assert.Contains(t, result, "SYSTEM_USER_CREATE=false")
		assert.Contains(t, result, "SYSTEM_USER_ENABLED=false")
		assert.Contains(t, result, "SYSTEM_USER_NAME=mod-inventory")
		assert.Contains(t, result, "SYSTEM_USER_USERNAME=mod-inventory")
	})

	t.Run("TestDisabledSystemUserEnv_EmptyEnv", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		moduleName := "test-module"

		// Act
		result := mv.DisabledSystemUserEnv(env, moduleName)

		// Assert
		assert.Len(t, result, 5)
		assert.Contains(t, result, "SYSTEM_USER_NAME=test-module")
	})
}

// ==================== KeycloakEnv Tests ====================

func TestKeycloakEnv(t *testing.T) {
	t.Run("TestKeycloakEnv_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigGlobalEnv: map[string]string{
				"kc_admin_client_id":     "admin-client",
				"kc_service_client_id":   "service-client",
				"kc_login_client_suffix": "-login",
			},
		}
		mv := moduleenv.New(act)
		env := []string{"EXISTING=value"}

		// Act
		result := mv.KeycloakEnv(env)

		// Assert
		assert.Len(t, result, 5)
		assert.Contains(t, result, "EXISTING=value")
		assert.Contains(t, result, "KC_URL=http://keycloak.eureka:8080")
		assert.Contains(t, result, "KC_ADMIN_CLIENT_ID=admin-client")
		assert.Contains(t, result, "KC_SERVICE_CLIENT_ID=service-client")
		assert.Contains(t, result, "KC_LOGIN_CLIENT_SUFFIX=-login")
	})

	t.Run("TestKeycloakEnv_EmptyGlobalEnv", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:            "test-action",
			ConfigGlobalEnv: map[string]string{},
		}
		mv := moduleenv.New(act)
		env := []string{}

		// Act
		result := mv.KeycloakEnv(env)

		// Assert
		assert.Len(t, result, 4)
		assert.Contains(t, result, "KC_URL=http://keycloak.eureka:8080")
	})
}

// ==================== ModuleEnv Tests ====================

func TestModuleEnv(t *testing.T) {
	t.Run("TestModuleEnv_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		envVars := []string{"EXISTING=value"}
		newEnv := map[string]any{
			"db_host":     "localhost",
			"db_port":     "5432",
			"db_username": "admin",
		}

		// Act
		result := mv.ModuleEnv(envVars, newEnv)

		// Assert
		assert.Len(t, result, 4)
		assert.Contains(t, result, "EXISTING=value")
		assert.Contains(t, result, "DB_HOST=localhost")
		assert.Contains(t, result, "DB_PORT=5432")
		assert.Contains(t, result, "DB_USERNAME=admin")
	})

	t.Run("TestModuleEnv_EmptyKey", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		envVars := []string{}
		newEnv := map[string]any{
			"":      "should-be-skipped",
			"VALID": "value",
		}

		// Act
		result := mv.ModuleEnv(envVars, newEnv)

		// Assert
		assert.Len(t, result, 1)
		assert.Contains(t, result, "VALID=value")
		for _, env := range result {
			assert.False(t, strings.HasPrefix(env, "="))
		}
	})

	t.Run("TestModuleEnv_EmptyNewEnv", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		envVars := []string{"EXISTING=value"}
		newEnv := map[string]any{}

		// Act
		result := mv.ModuleEnv(envVars, newEnv)

		// Assert
		assert.Len(t, result, 1)
		assert.Contains(t, result, "EXISTING=value")
	})

	t.Run("TestModuleEnv_UppercaseConversion", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		envVars := []string{}
		newEnv := map[string]any{
			"lowercase_key": "value1",
			"MixedCase":     "value2",
		}

		// Act
		result := mv.ModuleEnv(envVars, newEnv)

		// Assert
		assert.Len(t, result, 2)
		assert.Contains(t, result, "LOWERCASE_KEY=value1")
		assert.Contains(t, result, "MIXEDCASE=value2")
	})
}

// ==================== SidecarEnv Tests ====================

func TestSidecarEnv(t *testing.T) {
	t.Run("TestSidecarEnv_BothURLsEmpty", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{"EXISTING=value"}
		module := &models.ProxyModule{
			Metadata: models.ProxyModuleMetadata{
				Name:        "mod-inventory",
				Version:     helpers.StringPtr("1.0.0"),
				SidecarName: "mod-inventory-sidecar",
			},
		}
		privatePort := 8081
		moduleURL := ""
		sidecarURL := ""

		// Act
		result := mv.SidecarEnv(env, module, privatePort, moduleURL, sidecarURL)

		// Assert
		assert.Contains(t, result, "EXISTING=value")
		assert.Contains(t, result, "MODULE_NAME=mod-inventory")
		assert.Contains(t, result, "MODULE_VERSION=1.0.0")
		assert.Contains(t, result, "MODULE_URL=http://mod-inventory.eureka:8081")
		assert.Contains(t, result, "SIDECAR_NAME=mod-inventory-sidecar")
		assert.Contains(t, result, "SIDECAR_URL=http://mod-inventory-sidecar.eureka:8081")
	})

	t.Run("TestSidecarEnv_WithCustomURLs", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		module := &models.ProxyModule{
			Metadata: models.ProxyModuleMetadata{
				Name:        "mod-users",
				Version:     helpers.StringPtr("2.0.0"),
				SidecarName: "mod-users-sidecar",
			},
		}
		privatePort := 8082
		moduleURL := "http://custom-module:9000"
		sidecarURL := "http://custom-sidecar:9001"

		// Act
		result := mv.SidecarEnv(env, module, privatePort, moduleURL, sidecarURL)

		// Assert
		assert.Contains(t, result, "MODULE_NAME=mod-users")
		assert.Contains(t, result, "MODULE_VERSION=2.0.0")
		assert.Contains(t, result, "MODULE_URL=http://custom-module:9000")
		assert.Contains(t, result, "SIDECAR_NAME=mod-users-sidecar")
		assert.Contains(t, result, "SIDECAR_URL=http://custom-sidecar:9001")
	})

	t.Run("TestSidecarEnv_NonStandardPort", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		module := &models.ProxyModule{
			Metadata: models.ProxyModuleMetadata{
				Name:        "mod-circulation",
				Version:     helpers.StringPtr("1.5.0"),
				SidecarName: "mod-circulation-sidecar",
			},
		}
		privatePort := 9999
		moduleURL := ""
		sidecarURL := ""

		// Act
		result := mv.SidecarEnv(env, module, privatePort, moduleURL, sidecarURL)

		// Assert
		assert.Contains(t, result, "MODULE_URL=http://mod-circulation.eureka:9999")
		assert.Contains(t, result, "SIDECAR_URL=http://mod-circulation-sidecar.eureka:9999")
		assert.Contains(t, result, "QUARKUS_HTTP_PORT=9999")
	})

	t.Run("TestSidecarEnv_StandardPort8081", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}
		mv := moduleenv.New(act)
		env := []string{}
		module := &models.ProxyModule{
			Metadata: models.ProxyModuleMetadata{
				Name:        "mod-permissions",
				Version:     helpers.StringPtr("1.0.0"),
				SidecarName: "mod-permissions-sidecar",
			},
		}
		privatePort := 8081
		moduleURL := ""
		sidecarURL := ""

		// Act
		result := mv.SidecarEnv(env, module, privatePort, moduleURL, sidecarURL)

		// Assert
		assert.Contains(t, result, "MODULE_URL=http://mod-permissions.eureka:8081")
		// Should not contain QUARKUS_HTTP_PORT for standard port
		for _, envVar := range result {
			assert.False(t, strings.HasPrefix(envVar, "QUARKUS_HTTP_PORT="))
		}
	})
}
