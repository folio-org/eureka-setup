package moduleparams_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/moduleparams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Constructor Tests ====================

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		act := &action.Action{Name: "test-action"}

		// Act
		result := moduleparams.New(act)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, act, result.Action)
	})
}

// ==================== ReadBackendModulesFromConfig Tests ====================

func TestReadBackendModulesFromConfig_EmptyConfig(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_EmptyConfig_NoBackendModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                 "test-action",
			ConfigBackendModules: map[string]any{},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestReadBackendModulesFromConfig_Management(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_Management_FilterManagementModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mgr-tenants":   nil,
				"mgr-users":     nil,
				"mod-inventory": nil,
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(true, false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "mgr-tenants")
		assert.Contains(t, result, "mgr-users")
		assert.NotContains(t, result, "mod-inventory")
	})

	t.Run("TestReadBackendModulesFromConfig_Management_FilterNonManagementModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mgr-tenants":   nil,
				"mod-inventory": nil,
				"mod-users":     nil,
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "mod-inventory")
		assert.Contains(t, result, "mod-users")
		assert.NotContains(t, result, "mgr-tenants")
	})
}

func TestReadBackendModulesFromConfig_ConfigurableProperties(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithStringVersion", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVersionEntry: "1.0.0",
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "1.0.0", *module.ModuleVersion)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithNumericVersion", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVersionEntry: float64(2.5),
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "2.5", *module.ModuleVersion)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithCustomPort", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModulePortEntry: 9000,
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 9000, module.ModuleExposedServerPort)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithPrivatePort", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModulePrivatePortEntry: 8090,
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 8090, module.PrivatePort)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithDeployModuleFalse", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleDeployModuleEntry: false,
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 0, module.ModuleExposedServerPort)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithDeploySidecarFalse", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleDeploySidecarEntry: false,
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, 0, module.SidecarExposedServerPort)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithBooleanFlags", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.True(t, module.UseVault)
		assert.True(t, module.DisableSystemUser)
		assert.True(t, module.UseOkapiURL)
	})

	t.Run("TestReadBackendModulesFromConfig_ConfigurableProperties_WithEnvAndResources", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.NotNil(t, module.ModuleEnv)
		assert.NotNil(t, module.ModuleResources)
	})
}

func TestReadBackendModulesFromConfig_LocalDescriptor(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_LocalDescriptor_ValidLocalDescriptor", func(t *testing.T) {
		// Arrange
		tmpFile := filepath.Join(t.TempDir(), "descriptor.json")
		err := os.WriteFile(tmpFile, []byte(`{}`), 0600)
		require.NoError(t, err)

		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: tmpFile,
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Equal(t, tmpFile, module.LocalDescriptorPath)
	})

	t.Run("TestReadBackendModulesFromConfig_LocalDescriptor_InvalidLocalDescriptor", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:   "test-action",
			Params: &actionparams.ActionParams{},
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleLocalDescriptorPathEntry: "/nonexistent/path/descriptor.json",
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "local descriptor")
	})
}

func TestReadBackendModulesFromConfig_Volumes(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_Volumes_ValidVolume", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()

		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{tmpDir},
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Len(t, module.ModuleVolumes, 1)
		assert.Contains(t, module.ModuleVolumes, tmpDir)
	})

	t.Run("TestReadBackendModulesFromConfig_Volumes_VolumeWithEurekaVariable_Windows", func(t *testing.T) {
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
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{"$EUREKA/eureka-test-vol"},
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		if err == nil {
			require.Len(t, result, 1)
			module := result["mod-inventory"]
			assert.Len(t, module.ModuleVolumes, 1)
			assert.Contains(t, module.ModuleVolumes[0], homeDir)
		}
	})

	t.Run("TestReadBackendModulesFromConfig_Volumes_EmptyVolumes", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"mod-inventory": map[string]any{
					field.ModuleVolumesEntry: []any{},
				},
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["mod-inventory"]
		assert.Empty(t, module.ModuleVolumes)
	})
}

func TestReadBackendModulesFromConfig_EdgeModules(t *testing.T) {
	t.Run("TestReadBackendModulesFromConfig_EdgeModules_EdgeModuleNoSidecar", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                       "test-action",
			Params:                     &actionparams.ActionParams{},
			ReservedPorts:              []int{},
			ConfigApplicationPortStart: 8000,
			ConfigApplicationPortEnd:   9000,
			ConfigBackendModules: map[string]any{
				"edge-oai-pmh": nil,
			},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadBackendModulesFromConfig(false, false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["edge-oai-pmh"]
		assert.Equal(t, 0, module.SidecarExposedServerPort)
		assert.Equal(t, 0, module.SidecarExposedDebugPort)
	})
}

// ==================== ReadFrontendModulesFromConfig Tests ====================

func TestReadFrontendModulesFromConfig_EmptyConfig(t *testing.T) {
	t.Run("TestReadFrontendModulesFromConfig_EmptyConfig_NoFrontendModules", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestReadFrontendModulesFromConfig_DefaultProperties(t *testing.T) {
	t.Run("TestReadFrontendModulesFromConfig_DefaultProperties_NilValueCreatesDefault", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{"folio_inventory": nil},
			ConfigCustomFrontendModules: map[string]any{},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module, exists := result["folio_inventory"]
		assert.True(t, exists)
		assert.Equal(t, "folio_inventory", module.ModuleName)
		assert.True(t, module.DeployModule)
		assert.Nil(t, module.ModuleVersion)
	})

	t.Run("TestReadFrontendModulesFromConfig_DefaultProperties_WithVersion", func(t *testing.T) {
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.NotNil(t, module.ModuleVersion)
		assert.Equal(t, "2.0.0", *module.ModuleVersion)
	})

	t.Run("TestReadFrontendModulesFromConfig_DefaultProperties_WithDeployModuleFalse", func(t *testing.T) {
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.False(t, module.DeployModule)
	})

	t.Run("TestReadFrontendModulesFromConfig_DefaultProperties_WithLocalDescriptor", func(t *testing.T) {
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		require.Len(t, result, 1)
		module := result["folio_inventory"]
		assert.Equal(t, tmpFile, module.LocalDescriptorPath)
	})

	t.Run("TestReadFrontendModulesFromConfig_DefaultProperties_InvalidLocalDescriptor", func(t *testing.T) {
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
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "local descriptor")
	})
}

func TestReadFrontendModulesFromConfig_CustomModules(t *testing.T) {
	t.Run("TestReadFrontendModulesFromConfig_CustomModules_BothFrontendAndCustom", func(t *testing.T) {
		// Arrange
		act := &action.Action{
			Name:                        "test-action",
			ConfigFrontendModules:       map[string]any{"folio_inventory": nil},
			ConfigCustomFrontendModules: map[string]any{"custom_module": nil},
		}
		mp := moduleparams.New(act)

		// Act
		result, err := mp.ReadFrontendModulesFromConfig(false)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "folio_inventory")
		assert.Contains(t, result, "custom_module")
	})
}
