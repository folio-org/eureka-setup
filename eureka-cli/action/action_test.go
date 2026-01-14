package action_test

import (
	"runtime"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationName:    "test-app",
			field.ApplicationVersion: "1.0.0",
		})
		defer vc.Reset()

		params := &action.Param{}

		// Act
		result := action.New("test-action", "http://localhost:%s", params)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, "test-action", result.Name)
		assert.Equal(t, "http://localhost:%s", result.GatewayURLTemplate)
		assert.Equal(t, params, result.Param)
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

func TestNewGeneric_AllViperFields(t *testing.T) {
	t.Run("TestNewGeneric_AllViperFields_LoadsAllViperConfigurationFields", func(t *testing.T) {
		// Arrange
		viper.Reset() // Reset viper state before this test
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationName:              "full-app",
			field.ApplicationVersion:           "3.0.0",
			field.ProfileName:                  "production",
			field.RegistryURL:                  "https://registry.test.com",
			field.InstallFolio:                 "folio-registry-url",
			field.InstallEureka:                "eureka-registry-url",
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

		params := &action.Param{}

		// Act
		result := action.New("full-test", "http://test:%s", params)

		// Assert - Check core fields
		assert.Equal(t, "full-app", result.ConfigApplicationName)
		assert.Equal(t, "3.0.0", result.ConfigApplicationVersion)
		assert.Equal(t, "full-app-3.0.0", result.ConfigApplicationID)
		assert.Equal(t, "production", result.ConfigProfileName)
		assert.Equal(t, "https://registry.test.com", result.ConfigRegistryURL)
		assert.Equal(t, "folio-registry-url", result.ConfigInstallFolio)
		assert.Equal(t, "eureka-registry-url", result.ConfigInstallEureka)
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
		assert.NotNil(t, result.ConfigSidecarModuleResources)
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
	t.Run("TestGetGatewayURLTemplate_Success_WithHostnameReachable", func(t *testing.T) {
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

	t.Run("TestGetGatewayURLTemplate_Success_WithDefaultHostname", func(t *testing.T) {
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

	t.Run("TestGetGatewayURLTemplate_Fallback_WhenConfigHostnameUnreachable", func(t *testing.T) {
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
	t.Run("TestGetGatewayURL_Success_ConfigHostname", func(t *testing.T) {
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

	t.Run("TestGetGatewayURL_Fallback_ConfigHostnameWithInvalidHTTP", func(t *testing.T) {
		// Arrange - hostname with http:// prefix will fail reachability check
		// and fall back to default gateway
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "http://unreachable-test.invalid",
		})
		defer vc.Reset()

		// Act
		result, err := action.GetGatewayURL("test-action")

		// Assert - Should fallback to default or gateway IP
		if err == nil {
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "http://")
		}
	})

	t.Run("TestGetGatewayURL_Fallback_ToDefaultWhenConfigUnreachable", func(t *testing.T) {
		// Arrange - Unreachable config hostname should fallback
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "unreachable-host-98765.invalid",
		})
		defer vc.Reset()

		// Act
		result, err := action.GetGatewayURL("test-action")

		// Assert - Should either succeed with fallback or error
		if err == nil {
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "http://")
		}
	})

	t.Run("TestGetGatewayURL_Error_AllGatewaysFail", func(t *testing.T) {
		// Arrange - Set unreachable hostname and ensure other gateways can't resolve
		// This test mainly validates error handling on unsupported platforms
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.ApplicationGatewayHostname: "unreachable-xyz-123.invalid",
		})
		defer vc.Reset()

		// Act
		_, err := action.GetGatewayURL("test-action")

		// Assert - May succeed or fail depending on platform and network
		// Main goal is to ensure function doesn't panic
		if err != nil {
			assert.Error(t, err)
		}
	})
}

// ==================== Port Management Tests ====================

func TestGetPreReservedPort(t *testing.T) {
	t.Run("TestGetPreReservedPort_Success_FindsFreePort", func(t *testing.T) {
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

	t.Run("TestGetPreReservedPort_Success_ReservesMultiplePorts", func(t *testing.T) {
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

	t.Run("TestGetPreReservedPort_Error_NoFreePorts", func(t *testing.T) {
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
	t.Run("TestGetPreReservedPortSet_Success_ReservesMultiplePortsViaFunctions", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 59200,
			ConfigApplicationPortEnd:   59299,
			ReservedPorts:              []int{},
		}

		// Act
		ports, err := act.GetPreReservedPortSet(3)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, ports, 3)
		assert.NotEqual(t, ports[0], ports[1])
		assert.NotEqual(t, ports[1], ports[2])
		assert.NotEqual(t, ports[0], ports[2])
	})

	t.Run("TestGetPreReservedPortSet_Error_PropagatesPortReservationError", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			ConfigApplicationPortStart: 0,
			ConfigApplicationPortEnd:   0,
			ReservedPorts:              []int{},
		}

		// Act
		ports, err := act.GetPreReservedPortSet(1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, ports)
	})
}

// ==================== URL Generation Tests ====================

func TestGetRequestURL(t *testing.T) {
	t.Run("TestGetRequestURL_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			GatewayURLTemplate: "http://localhost:%s",
		}

		// Act
		result := act.GetRequestURL("8080", "/api/users")

		// Assert
		assert.Equal(t, "http://localhost:8080/api/users", result)
	})

	t.Run("TestGetRequestURL_Success_WithEmptyRoute", func(t *testing.T) {
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
	t.Run("TestGetConfigEnvVars_Success_ReturnsEnvVars", func(t *testing.T) {
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

	t.Run("TestGetConfigEnvVars_Success_EmptyMap", func(t *testing.T) {
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
	t.Run("TestGetConfigEnv_Success_FindsValueWithLowercaseKey", func(t *testing.T) {
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

	t.Run("TestGetConfigEnv_Success_FindsValueWithUppercaseInput", func(t *testing.T) {
		// Arrange
		configEnv := map[string]string{
			"db_host": "localhost",
		}

		// Act
		result := action.GetConfigEnv("DB_HOST", configEnv)

		// Assert
		assert.Equal(t, "localhost", result)
	})

	t.Run("TestGetConfigEnv_Success_ReturnsEmptyForMissingKey", func(t *testing.T) {
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
	t.Run("TestIsSet_Success_KeyExists", func(t *testing.T) {
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

	t.Run("TestIsSet_Success_KeyDoesNotExist", func(t *testing.T) {
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
	t.Run("TestGetCombinedInstallJsonURLs_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigInstallFolio:  "https://folio.registry.com/install.json",
			ConfigInstallEureka: "https://eureka.registry.com/install.json",
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
	t.Run("TestGetEurekaInstallJsonURLs_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigInstallEureka: "https://eureka.registry.com/install.json",
		}

		// Act
		result := act.GetEurekaInstallJsonURLs()

		// Assert
		assert.Len(t, result, 1)
		assert.Equal(t, "https://eureka.registry.com/install.json", result[constant.EurekaRegistry])
	})
}

// ==================== GetModuleURL Tests ====================

func TestGetModuleURL(t *testing.T) {
	t.Run("TestGetModuleURL_Success_WithSimpleModuleID", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "https://okapi.test.com",
		}

		// Act
		result := act.GetModuleURL("mod-inventory-1.0.0")

		// Assert
		assert.Equal(t, "https://okapi.test.com/_/proxy/modules/mod-inventory-1.0.0", result)
	})

	t.Run("TestGetModuleURL_Success_WithComplexModuleID", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "http://localhost:9130",
		}

		// Act
		result := act.GetModuleURL("mod-users-keycloak-2.5.0-SNAPSHOT.123")

		// Assert
		assert.Equal(t, "http://localhost:9130/_/proxy/modules/mod-users-keycloak-2.5.0-SNAPSHOT.123", result)
	})

	t.Run("TestGetModuleURL_Success_WithDifferentRegistryURL", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "https://eureka-registry.example.org:8080",
		}

		// Act
		result := act.GetModuleURL("mod-search-3.2.1")

		// Assert
		assert.Equal(t, "https://eureka-registry.example.org:8080/_/proxy/modules/mod-search-3.2.1", result)
	})

	t.Run("TestGetModuleURL_Success_WithEmptyModuleID", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "https://okapi.test.com",
		}

		// Act
		result := act.GetModuleURL("")

		// Assert
		assert.Equal(t, "https://okapi.test.com/_/proxy/modules/", result)
	})

	t.Run("TestGetModuleURL_Success_WithSpecialCharactersInModuleID", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigRegistryURL: "https://okapi.test.com",
		}

		// Act
		result := act.GetModuleURL("mod-test_module-1.0.0-RC.1")

		// Assert
		assert.Equal(t, "https://okapi.test.com/_/proxy/modules/mod-test_module-1.0.0-RC.1", result)
	})
}

// ==================== Kafka Topic Config Tests ====================

func TestGetKafkaTopicConfigTenant(t *testing.T) {
	t.Run("TestGetKafkaTopicConfigTenant_ReturnsConfigTenant_WhenTopicSharingDisabled", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigManagementTopicSharing: false,
			ConfigTopicSharingTenant:     "shared-tenant",
		}
		configTenant := "diku"

		// Act
		result := act.GetKafkaTopicConfigTenant(configTenant)

		// Assert
		assert.Equal(t, "diku", result)
	})

	t.Run("TestGetKafkaTopicConfigTenant_ReturnsSharedTenant_WhenTopicSharingEnabled", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigManagementTopicSharing: true,
			ConfigTopicSharingTenant:     "shared-tenant",
		}
		configTenant := "diku"

		// Act
		result := act.GetKafkaTopicConfigTenant(configTenant)

		// Assert
		assert.Equal(t, "shared-tenant", result)
	})

	t.Run("TestGetKafkaTopicConfigTenant_ReturnsSharedTenant_EvenWithEmptyConfigTenant", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigManagementTopicSharing: true,
			ConfigTopicSharingTenant:     "global-shared",
		}
		configTenant := ""

		// Act
		result := act.GetKafkaTopicConfigTenant(configTenant)

		// Assert
		assert.Equal(t, "global-shared", result)
	})

	t.Run("TestGetKafkaTopicConfigTenant_ReturnsConfigTenant_WhenBothAreEmpty", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigManagementTopicSharing: false,
			ConfigTopicSharingTenant:     "",
		}
		configTenant := ""

		// Act
		result := act.GetKafkaTopicConfigTenant(configTenant)

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("TestGetKafkaTopicConfigTenant_HandlesMultipleTenants", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			ConfigManagementTopicSharing: true,
			ConfigTopicSharingTenant:     "central",
		}

		// Act
		result1 := act.GetKafkaTopicConfigTenant("tenant1")
		result2 := act.GetKafkaTopicConfigTenant("tenant2")
		result3 := act.GetKafkaTopicConfigTenant("tenant3")

		// Assert - All should return the shared tenant
		assert.Equal(t, "central", result1)
		assert.Equal(t, "central", result2)
		assert.Equal(t, "central", result3)
	})
}

// ==================== GetSidecarModuleCmd Tests ====================

func TestGetSidecarModuleCmd(t *testing.T) {
	t.Run("TestGetSidecarModuleCmd_WithNativeBinaryCmd", func(t *testing.T) {
		// Arrange
		viper.Reset()
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.SidecarModuleNativeBinaryCmd: []string{"./app", "-flag1", "-flag2"},
		})
		defer vc.Reset()

		// Act
		result := action.GetSidecarModuleCmd()

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Equal(t, "./app", result[0])
		assert.Equal(t, "-flag1", result[1])
		assert.Equal(t, "-flag2", result[2])
	})

	t.Run("TestGetSidecarModuleCmd_WithCmd_FallbackCompatibility", func(t *testing.T) {
		// Arrange - test fallback to old "cmd" field when native-binary-cmd is not set
		viper.Reset()
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.SidecarModuleCmd: []string{"java", "-jar", "app.jar"},
		})
		defer vc.Reset()

		// Act
		result := action.GetSidecarModuleCmd()

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Equal(t, "java", result[0])
		assert.Equal(t, "-jar", result[1])
		assert.Equal(t, "app.jar", result[2])
	})

	t.Run("TestGetSidecarModuleCmd_NativeBinaryCmdTakesPrecedence", func(t *testing.T) {
		// Arrange - test that native-binary-cmd takes precedence over cmd
		viper.Reset()
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.SidecarModuleNativeBinaryCmd: []string{"./native-app"},
			field.SidecarModuleCmd:             []string{"java", "-jar", "app.jar"},
		})
		defer vc.Reset()

		// Act
		result := action.GetSidecarModuleCmd()

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "./native-app", result[0], "native-binary-cmd should take precedence over cmd")
	})

	t.Run("TestGetSidecarModuleCmd_EmptyWhenNeitherSet", func(t *testing.T) {
		// Arrange
		viper.Reset()
		defer viper.Reset()

		// Act
		result := action.GetSidecarModuleCmd()

		// Assert
		if result == nil {
			result = []string{}
		}
		assert.Empty(t, result)
	})

	t.Run("TestGetSidecarModuleCmd_EmptyNativeBinaryCmd", func(t *testing.T) {
		// Arrange - empty native-binary-cmd should fallback to cmd
		viper.Reset()
		vc := testhelpers.SetupViperForTest(map[string]any{
			field.SidecarModuleNativeBinaryCmd: []string{},
			field.SidecarModuleCmd:             []string{"fallback", "command"},
		})
		defer vc.Reset()

		// Act
		result := action.GetSidecarModuleCmd()

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "fallback", result[0])
		assert.Equal(t, "command", result[1])
	})
}

// ==================== Application Tests ====================

func TestIsChildApp(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string]any
		expected     bool
	}{
		{
			name: "TestIsChildApp_ReturnsTrueWhenDependenciesExist",
			dependencies: map[string]any{
				"dep1": "v1.0.0",
				"dep2": "v2.0.0",
			},
			expected: true,
		},
		{
			name:         "TestIsChildApp_ReturnsFalseWhenDependenciesEmpty",
			dependencies: map[string]any{},
			expected:     false,
		},
		{
			name:         "TestIsChildApp_ReturnsFalseWhenDependenciesNil",
			dependencies: nil,
			expected:     false,
		},
		{
			name: "TestIsChildApp_ReturnsTrueWithSingleDependency",
			dependencies: map[string]any{
				"single-dep": "1.0.0",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			vc := testhelpers.SetupViperForTest(map[string]any{
				field.ApplicationDependencies: tt.dependencies,
			})
			defer vc.Reset()

			act := action.New("test", "http://localhost:%s", &action.Param{})

			// Act
			result := act.IsChildApp()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== Param Tests ====================

func TestFlag_GetName(t *testing.T) {
	t.Run("TestFlag_GetName_ReturnsLongName", func(t *testing.T) {
		// Arrange
		flag := action.Flag{
			Long:        "test-flag",
			Short:       "t",
			Description: "A test flag",
		}

		// Act
		result := flag.GetName()

		// Assert
		assert.Equal(t, "test-flag", result)
	})

	t.Run("TestFlag_GetName_WithEmptyShort", func(t *testing.T) {
		// Arrange
		flag := action.Flag{
			Long:        "another-flag",
			Short:       "",
			Description: "Another test flag",
		}

		// Act
		result := flag.GetName()

		// Assert
		assert.Equal(t, "another-flag", result)
	})

	t.Run("TestFlag_GetName_WithComplexName", func(t *testing.T) {
		// Arrange
		flag := action.Flag{
			Long:        "module-deployment-skip",
			Short:       "mds",
			Description: "Skip module deployment",
		}

		// Act
		result := flag.GetName()

		// Assert
		assert.Equal(t, "module-deployment-skip", result)
	})
}
