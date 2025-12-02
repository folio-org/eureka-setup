package moduleprops_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/moduleprops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}

		// Act
		result := moduleprops.New(act)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, act, result.Action)
	})
}

func TestReadBackendModules_PortExhaustion(t *testing.T) {
	t.Run("TestReadBackendModules_PortExhaustion_NoAvailablePorts", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   8000, // No range
			ConfigBackendModules: map[string]any{
				"mod-inventory": nil,
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "port")
	})
}

// ==================== ReadBackendModulesFromConfig Tests ====================

func TestReadBackendModules_EmptyConfig(t *testing.T) {
	t.Run("TestReadBackendModules_EmptyConfig_NoBackendModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                 "test-action",
			ConfigBackendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestReadBackendModules_Management(t *testing.T) {
	t.Run("TestReadBackendModules_Management_FilterManagementModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mgr-tenants":   nil,
				"mgr-users":     nil,
				"mod-inventory": nil,
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(true, false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "mgr-tenants")
		assert.Contains(t, result, "mgr-users")
		assert.NotContains(t, result, "mod-inventory")
	})

	t.Run("TestReadBackendModules_Management_FilterNonManagementModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mgr-tenants":   nil,
				"mod-inventory": nil,
				"mod-users":     nil,
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "mod-inventory")
		assert.Contains(t, result, "mod-users")
		assert.NotContains(t, result, "mgr-tenants")
	})
}

func TestReadBackendModules_ConfigurableProperties(t *testing.T) {
	t.Run("TestReadBackendModules_ConfigurableProperties_WithStringVersion", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVersionEntry: "1.0.0",
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "1.0.0", *module.ModuleVersion)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithNumericVersion", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVersionEntry: float64(2.5),
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "2.5", *module.ModuleVersion)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithCustomPort", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModulePortEntry: 9000,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 9000, module.ModuleExposedServerPort)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithPrivatePort", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModulePrivatePortEntry: 8090,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 8090, module.PrivatePort)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithDeployModuleFalse", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleDeployModuleEntry: false,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 0, module.ModuleExposedServerPort)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithDeploySidecarFalse", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleDeploySidecarEntry: false,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 0, module.SidecarExposedServerPort)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithBooleanFlags", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleUseVaultEntry:          true,
					field.ModuleDisableSystemUserEntry: true,
					field.ModuleUseOkapiURLEntry:       true,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.True(t, module.UseVault)
		assert.True(t, module.DisableSystemUser)
		assert.True(t, module.UseOkapiURL)
	})

	t.Run("TestReadBackendModules_ConfigurableProperties_WithEnvAndResources", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleEnvEntry: map[string]any{
						"DB_HOST": "localhost",
					},
					field.ModuleResourceEntry: map[string]any{
						"memory": "512m",
					},
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleEnv)
		assert.NotNil(t, module.ModuleResources)
	})
}

func TestReadBackendModules_LocalDescriptor(t *testing.T) {
	t.Run("TestReadBackendModules_LocalDescriptor_ValidLocalDescriptor", func(t *testing.T) {
		// Arrange
		tmpFile := filepath.Join(t.TempDir(), "descriptor.json")
		err := os.WriteFile(tmpFile, []byte(`{}`), 0600)
		require.NoError(t, err)

		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: tmpFile,
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, tmpFile, module.LocalDescriptorPath)
	})

	t.Run("TestReadBackendModules_LocalDescriptor_InvalidLocalDescriptor", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:  "test-action",
			Param: &action.Param{},
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: "/nonexistent/path/descriptor.json",
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "local descriptor")
	})
}

func TestReadBackendModules_Volumes(t *testing.T) {
	t.Run("TestReadBackendModules_Volumes_ValidVolume", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{tmpDir},
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Len(t, module.ModuleVolumes, 1)
		assert.Contains(t, module.ModuleVolumes, tmpDir)
	})

	t.Run("TestReadBackendModules_Volumes_VolumeWithEurekaVariable_Windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Skipping Windows-specific test")
		}

		// Arrange
		homeDir, err := helpers.GetHomeDirPath()
		require.NoError(t, err)

		// Create test directory
		testPath := filepath.Join(homeDir, "eureka-test-vol")
		err = os.MkdirAll(testPath, 0750)
		if err != nil {
			t.Skip("Cannot create test directory")
		}
		defer func() { _ = os.RemoveAll(testPath) }()

		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{"$EUREKA/eureka-test-vol"},
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		if err == nil {
			require.Len(t, result, 1)
			module := result["mod-inventory"]
			assert.Len(t, module.ModuleVolumes, 1)
			assert.Contains(t, module.ModuleVolumes[0], homeDir)
		}
	})

	t.Run("TestReadBackendModules_Volumes_EmptyVolumes", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{},
				},
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Empty(t, module.ModuleVolumes)
	})
}

func TestReadBackendModules_EdgeModules(t *testing.T) {
	t.Run("TestReadBackendModules_EdgeModules_EdgeModuleNoSidecar", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Param:                      &action.Param{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"edge-oai-pmh": nil,
			},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadBackendModules(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["edge-oai-pmh"]
		assert.Equal(t, 0, module.SidecarExposedServerPort)
		assert.Equal(t, 0, module.SidecarExposedDebugPort)
	})
}

// ==================== ReadFrontendModules Tests ====================

func TestReadFrontendModules_EmptyConfig(t *testing.T) {
	t.Run("TestReadFrontendModules_EmptyConfig_NoFrontendModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestReadFrontendModules_DefaultProperties(t *testing.T) {
	t.Run("TestReadFrontendModules_DefaultProperties_NilValueCreatesDefault", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{"folio_inventory": nil},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module, exists := result["folio_inventory"]
		assert.True(t, exists)
		assert.Equal(t, "folio_inventory", module.ModuleName)
		assert.True(t, module.DeployModule)
		assert.Nil(t, module.ModuleVersion)
	})

	t.Run("TestReadFrontendModules_DefaultProperties_WithVersion", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigFrontendModules: map[string]any{
				"folio_inventory": map[string]any{
					field.ModuleVersionEntry: "2.0.0",
				},
			},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "2.0.0", *module.ModuleVersion)
	})

	t.Run("TestReadFrontendModules_DefaultProperties_WithDeployModuleFalse", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigFrontendModules: map[string]any{
				"folio_inventory": map[string]any{
					field.ModuleDeployModuleEntry: false,
				},
			},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.False(t, module.DeployModule)
	})

	t.Run("TestReadFrontendModules_DefaultProperties_WithLocalDescriptor", func(t *testing.T) {
		// Arrange
		tmpFile := filepath.Join(t.TempDir(), "frontend-descriptor.json")
		err := os.WriteFile(tmpFile, []byte(`{}`), 0600)
		require.NoError(t, err)

		act := &action.Action{
			Name: "test-action",
			ConfigFrontendModules: map[string]any{
				"folio_inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: tmpFile,
				},
			},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.Equal(t, tmpFile, module.LocalDescriptorPath)
	})

	t.Run("TestReadFrontendModules_DefaultProperties_InvalidLocalDescriptor", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name: "test-action",
			ConfigFrontendModules: map[string]any{
				"folio_inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: "/nonexistent/frontend.json",
				},
			},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "local descriptor")
	})
}

func TestReadFrontendModules_CustomModules(t *testing.T) {
	t.Run("TestReadFrontendModules_CustomModules_BothFrontendAndCustom", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{"folio_inventory": nil},
			ConfigCustomFrontendModules: map[string]any{"custom_module": nil},
		}
		mp := moduleprops.New(act)

		// Act
		result, err := mp.ReadFrontendModules(false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "folio_inventory")
		assert.Contains(t, result, "custom_module")
	})
}
