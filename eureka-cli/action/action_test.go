package action_test

import (
	"runtime"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/internal/testhelpers"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationName:    "test-app",
			field.ApplicationVersion: "1.0.0",
		})
		defer vc.Reset()

		params := &actionparams.ActionParams{}

		// Act
		result := action.New("test-action", "http://localhost:%s", params)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, "test-action", result.Name)
		assert.Equal(t, "http://localhost:%s", result.GatewayURLTemplate)
		assert.Equal(t, params, result.Params)
		assert.Equal(t, "", result.VaultRootToken)
		assert.Equal(t, "", result.KeycloakAccessToken)
		assert.Equal(t, "", result.KeycloakMasterAccessToken)
		assert.Equal(t, "test-app", result.ConfigApplicationName)
		assert.Equal(t, "1.0.0", result.ConfigApplicationVersion)
		assert.Equal(t, "test-app-1.0.0", result.ConfigApplicationID)
		assert.NotNil(t, result.Caser)
		assert.NotNil(t, result.ReservedPorts)
		assert.Empty(t, result.ReservedPorts)
	})
}

func TestNewWithCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationName:    "test-app",
			field.ApplicationVersion: "2.0.0",
			field.ProfileName:        "test-profile",
		})
		defer vc.Reset()

		params := &actionparams.ActionParams{}

		// Act
		result := action.NewWithCredentials(
			"test-action",
			"http://localhost:%s",
			params,
			"vault-token-123",
			"keycloak-token-456",
			"master-token-789",
		)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, "test-action", result.Name)
		assert.Equal(t, "http://localhost:%s", result.GatewayURLTemplate)
		assert.Equal(t, params, result.Params)
		assert.Equal(t, "vault-token-123", result.VaultRootToken)
		assert.Equal(t, "keycloak-token-456", result.KeycloakAccessToken)
		assert.Equal(t, "master-token-789", result.KeycloakMasterAccessToken)
		assert.Equal(t, "test-app", result.ConfigApplicationName)
		assert.Equal(t, "2.0.0", result.ConfigApplicationVersion)
		assert.Equal(t, "test-profile", result.ConfigProfile)
	})
}

func TestNewGeneric_AllViperFields(t *testing.T) {
	t.Run("LoadsAllViperConfigurationFields", func(t *testing.T) {
		// Arrange
		viper.Reset() // Reset viper state before this test
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationName:              "full-app",
			field.ApplicationVersion:           "3.0.0",
			field.ProfileName:                  "production",
			field.RegistryURL:                  "https://registry.test.com",
			field.InstallFolio:                 "folio-registry-url",
			field.InstallEureka:                "eureka-registry-url",
			field.ApplicationPlatform:          "kubernetes",
			field.ApplicationFetchDescriptors:  true,
			field.ApplicationPortStart:         8000,
			field.ApplicationPortEnd:           9000,
			field.ApplicationStripesBranch:     "main",
			field.ApplicationGatewayHostname:   "gateway.test.com",
			field.NamespacesPlatformCompleteUI: "platform-ui-namespace",
			field.ApplicationDependencies: map[string]any{
				"dep1": "v1",
			},
			field.Env: map[string]any{
				"VAR1": "value1",
			},
			field.SidecarModule: map[string]any{
				"name": "sidecar-module",
			},
			field.SidecarModuleResources: map[string]any{
				"memory": "512m",
			},
			field.BackendModules: map[string]any{
				"mod-inventory": nil,
			},
			field.FrontendModules: map[string]any{
				"folio_inventory": nil,
			},
			field.CustomFrontendModules: map[string]any{
				"custom-ui": nil,
			},
			field.Tenants: map[string]any{
				"diku": nil,
			},
			field.Roles: map[string]any{
				"admin": nil,
			},
			field.Users: map[string]any{
				"testuser": nil,
			},
			field.RolesCapabilitySetsEntry: map[string]any{
				"admin-caps": nil,
			},
			field.Consortiums: map[string]any{
				"test-consortium": nil,
			},
		})
		defer func() {
			vc.Reset()
			viper.Reset()
		}()

		params := &actionparams.ActionParams{}

		// Act
		result := action.New("full-test", "http://test:%s", params)

		// Assert - Check core fields
		assert.Equal(t, "full-app", result.ConfigApplicationName)
		assert.Equal(t, "3.0.0", result.ConfigApplicationVersion)
		assert.Equal(t, "full-app-3.0.0", result.ConfigApplicationID)
		assert.Equal(t, "production", result.ConfigProfile)
		assert.Equal(t, "https://registry.test.com", result.ConfigRegistryURL)
		assert.Equal(t, "folio-registry-url", result.ConfigFolioRegistry)
		assert.Equal(t, "eureka-registry-url", result.ConfigEurekaRegistry)
		assert.Equal(t, "kubernetes", result.ConfigApplicationPlatform)
		assert.True(t, result.ConfigApplicationFetchDescriptors)
		assert.Equal(t, 8000, result.ConfigApplicationPortStart)
		assert.Equal(t, 9000, result.ConfigApplicationPortEnd)
		assert.Equal(t, "main", result.ConfigApplicationStripesBranch)
		assert.Equal(t, "gateway.test.com", result.ConfigApplicationGatewayHostname)
		assert.Equal(t, "platform-ui-namespace", result.ConfigNamespacePlatformCompleteUI)

		// Assert - Check maps are loaded
		assert.NotNil(t, result.ConfigApplicationDependencies)
		assert.NotNil(t, result.ConfigGlobalEnv)
		assert.NotNil(t, result.ConfigSidecarModule)
		assert.NotNil(t, result.ConfigSidecarResources)
		assert.NotNil(t, result.ConfigBackendModules)
		assert.NotNil(t, result.ConfigFrontendModules)
		assert.NotNil(t, result.ConfigCustomFrontendModules)
		assert.NotNil(t, result.ConfigTenants)
		assert.NotNil(t, result.ConfigRoles)
		assert.NotNil(t, result.ConfigUsers)
		assert.NotNil(t, result.ConfigRolesCapabilitySets)
		assert.NotNil(t, result.ConfigConsortiums)
	})
}

// ==================== GetGatewayURLTemplate Tests ====================

func TestGetGatewayURLTemplate(t *testing.T) {
	t.Run("Success_WithHostnameReachable", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "localhost",
		})
		defer vc.Reset()

		// Act
		result, err := action.GetGatewayURLTemplate("test-action")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:%s", result)
	})

	t.Run("Success_WithDefaultHostname", func(t *testing.T) {
		// Arrange
		viper.Reset()

		// Act
		result, err := action.GetGatewayURLTemplate("test-action")

		// Assert
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			// On Linux/Darwin, should fall back to docker gateway IP
			assert.NoError(t, err)
			assert.Contains(t, result, ":%s")
		} else {
			// On other platforms, may succeed or fail depending on hostname.docker.internal availability
			// Just check it returns something valid or an error
			if err == nil {
				assert.Contains(t, result, ":%s")
			} else {
				assert.Error(t, err)
			}
		}
	})

	t.Run("Fallback_WhenConfigHostnameUnreachable", func(t *testing.T) {
		// Arrange - Set unreachable hostname, should fall back to default
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "unreachable-host-12345.invalid",
		})
		defer vc.Reset()

		// Act
		result, err := action.GetGatewayURLTemplate("test-action")

		// Assert - Falls back to default hostname or gateway IP
		if err == nil {
			assert.Contains(t, result, ":%s")
		} else {
			assert.Error(t, err)
		}
	})
}

// ==================== GetGatewayURL Tests ====================

func TestGetGatewayURL(t *testing.T) {
	t.Run("Success_ConfigHostname", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "localhost",
		})
		defer vc.Reset()

		// Act
		result, err := action.GetGatewayURL("test-action")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost", result)
	})
}

// ==================== Port Management Tests ====================

func TestGetPreReservedPort(t *testing.T) {
	t.Run("Success_FindsFreePort", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 59000,
			ConfigApplicationPortEnd:   59999,
			ReservedPorts:              []int{},
		}

		// Act
		port, err := act.GetPreReservedPort()

		// Assert
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, port, 59000)
		assert.LessOrEqual(t, port, 59999)
		assert.Contains(t, act.ReservedPorts, port)
	})

	t.Run("Success_ReservesMultiplePorts", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 59100,
			ConfigApplicationPortEnd:   59199,
			ReservedPorts:              []int{},
		}

		// Act
		port1, err1 := act.GetPreReservedPort()
		port2, err2 := act.GetPreReservedPort()
		port3, err3 := act.GetPreReservedPort()

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NoError(t, err3)
		assert.NotEqual(t, port1, port2)
		assert.NotEqual(t, port2, port3)
		assert.NotEqual(t, port1, port3)
		assert.Len(t, act.ReservedPorts, 3)
		assert.Contains(t, act.ReservedPorts, port1)
		assert.Contains(t, act.ReservedPorts, port2)
		assert.Contains(t, act.ReservedPorts, port3)
	})

	t.Run("Error_NoFreePorts", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 0,
			ConfigApplicationPortEnd:   0,
			ReservedPorts:              []int{},
		}

		// Act
		port, err := act.GetPreReservedPort()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, port)
		assert.Contains(t, err.Error(), "failed to find free TCP ports")
	})
}

func TestGetPreReservedPortSet(t *testing.T) {
	t.Run("Success_ReservesMultiplePortsViaFunctions", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 59200,
			ConfigApplicationPortEnd:   59299,
			ReservedPorts:              []int{},
		}

		fns := []func() (int, error){
			act.GetPreReservedPort,
			act.GetPreReservedPort,
			act.GetPreReservedPort,
		}

		// Act
		ports, err := act.GetPreReservedPortSet(fns)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, ports, 3)
		assert.NotEqual(t, ports[0], ports[1])
		assert.NotEqual(t, ports[1], ports[2])
		assert.NotEqual(t, ports[0], ports[2])
	})

	t.Run("Error_PropagatesPortReservationError", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 0,
			ConfigApplicationPortEnd:   0,
			ReservedPorts:              []int{},
		}

		fns := []func() (int, error){
			act.GetPreReservedPort,
		}

		// Act
		ports, err := act.GetPreReservedPortSet(fns)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, ports)
	})
}

// ==================== URL Generation Tests ====================

func TestGetRequestURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			GatewayURLTemplate: "http://localhost:%s",
		}

		// Act
		result := act.GetRequestURL("8080", "/api/users")

		// Assert
		assert.Equal(t, "http://localhost:8080/api/users", result)
	})

	t.Run("Success_WithEmptyRoute", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			GatewayURLTemplate: "http://test:%s",
		}

		// Act
		result := act.GetRequestURL("9000", "")

		// Assert
		assert.Equal(t, "http://test:9000", result)
	})
}

// ==================== Environment Variable Tests ====================

func TestGetConfigEnvVars(t *testing.T) {
	t.Run("Success_ReturnsEnvVars", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			"test-env": map[string]any{
				"db_host": "localhost",
				"db_port": "5432",
			},
		})
		defer vc.Reset()

		act := &action.Action{Name: "test-action"}

		// Act
		result := act.GetConfigEnvVars("test-env")

		// Assert
		assert.Len(t, result, 2)
		assert.Contains(t, result, "DB_HOST=localhost")
		assert.Contains(t, result, "DB_PORT=5432")
	})

	t.Run("Success_EmptyMap", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			"empty-env": map[string]any{},
		})
		defer vc.Reset()

		act := &action.Action{Name: "test-action"}

		// Act
		result := act.GetConfigEnvVars("empty-env")

		// Assert
		assert.Empty(t, result)
	})
}

func TestGetConfigEnv(t *testing.T) {
	t.Run("Success_FindsValueWithLowercaseKey", func(t *testing.T) {
		// Arrange
		configEnv := map[string]string{
			"db_host": "localhost",
			"db_port": "5432",
		}

		// Act
		result := action.GetConfigEnv("db_host", configEnv)

		// Assert
		assert.Equal(t, "localhost", result)
	})

	t.Run("Success_FindsValueWithUppercaseInput", func(t *testing.T) {
		// Arrange
		configEnv := map[string]string{
			"db_host": "localhost",
		}

		// Act
		result := action.GetConfigEnv("DB_HOST", configEnv)

		// Assert
		assert.Equal(t, "localhost", result)
	})

	t.Run("Success_ReturnsEmptyForMissingKey", func(t *testing.T) {
		// Arrange
		configEnv := map[string]string{
			"db_host": "localhost",
		}

		// Act
		result := action.GetConfigEnv("missing_key", configEnv)

		// Assert
		assert.Equal(t, "", result)
	})
}

// ==================== Configuration Check Tests ====================

func TestIsSet(t *testing.T) {
	t.Run("Success_KeyExists", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			"test-key": "test-value",
		})
		defer vc.Reset()

		// Act
		result := action.IsSet("test-key")

		// Assert
		assert.True(t, result)
	})

	t.Run("Success_KeyDoesNotExist", func(t *testing.T) {
		// Arrange
		viper.Reset()

		// Act
		result := action.IsSet("nonexistent-key")

		// Assert
		assert.False(t, result)
	})
}

// ==================== Registry URL Tests ====================

func TestGetCombinedInstallJsonURLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigFolioRegistry:  "https://folio.registry.com/install.json",
			ConfigEurekaRegistry: "https://eureka.registry.com/install.json",
		}

		// Act
		result := act.GetCombinedInstallJsonURLs()

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "https://folio.registry.com/install.json", result[constant.FolioRegistry])
		assert.Equal(t, "https://eureka.registry.com/install.json", result[constant.EurekaRegistry])
	})
}

func TestGetEurekaInstallJsonURLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigEurekaRegistry: "https://eureka.registry.com/install.json",
		}

		// Act
		result := act.GetEurekaInstallJsonURLs()

		// Assert
		assert.Len(t, result, 1)
		assert.Equal(t, "https://eureka.registry.com/install.json", result[constant.EurekaRegistry])
	})
}

func TestGetCombinedRegistryURLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "https://registry.test.com",
		}

		// Act
		result := act.GetCombinedRegistryURLs()

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "https://registry.test.com", result[constant.FolioRegistry])
		assert.Equal(t, "https://registry.test.com", result[constant.EurekaRegistry])
	})
}
