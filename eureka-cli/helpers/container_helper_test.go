package helpers

import (
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/viper"
)

func TestNewModuleNetworkConfig(t *testing.T) {
	config := NewModuleNetworkConfig()

	if config == nil {
		t.Fatal("NewModuleNetworkConfig() returned nil")
	}

	if config.EndpointsConfig == nil {
		t.Fatal("EndpointsConfig is nil")
	}

	endpointConfig, exists := config.EndpointsConfig[constant.NetworkID]
	if !exists {
		t.Fatalf("Expected endpoint config for network ID %s", constant.NetworkID)
	}

	if endpointConfig.NetworkID != constant.NetworkID {
		t.Errorf("NetworkID = %s, want %s", endpointConfig.NetworkID, constant.NetworkID)
	}

	if len(endpointConfig.Aliases) == 0 {
		t.Error("Expected at least one alias")
	} else if endpointConfig.Aliases[0] != constant.NetworkAlias {
		t.Errorf("First alias = %s, want %s", endpointConfig.Aliases[0], constant.NetworkAlias)
	}
}

func TestCreateExposedPorts(t *testing.T) {
	tests := []struct {
		name       string
		serverPort int
	}{
		{
			name:       "standard port 8080",
			serverPort: 8080,
		},
		{
			name:       "custom port 9000",
			serverPort: 9000,
		},
		{
			name:       "port 80",
			serverPort: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portSet := CreateExposedPorts(tt.serverPort)

			if portSet == nil {
				t.Fatal("CreateExposedPorts() returned nil")
			}

			// Check if server port is exposed
			serverPortKey := nat.Port(strconv.Itoa(tt.serverPort))
			if _, exists := (*portSet)[serverPortKey]; !exists {
				t.Errorf("Server port %d not found in exposed ports", tt.serverPort)
			}

			// Check if debug port is exposed
			debugPortKey := nat.Port(constant.DebugPort)
			if _, exists := (*portSet)[debugPortKey]; !exists {
				t.Errorf("Debug port %s not found in exposed ports", constant.DebugPort)
			}

			// Should have exactly 2 ports
			if len(*portSet) != 2 {
				t.Errorf("Expected 2 exposed ports, got %d", len(*portSet))
			}
		})
	}
}

func TestCreatePortBindings(t *testing.T) {
	tests := []struct {
		name                string
		hostServerPort      int
		hostServerDebugPort int
		serverPort          int
	}{
		{
			name:                "standard binding",
			hostServerPort:      8080,
			hostServerDebugPort: 5005,
			serverPort:          8080,
		},
		{
			name:                "custom ports",
			hostServerPort:      9000,
			hostServerDebugPort: 9001,
			serverPort:          8081,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portMap := CreatePortBindings(tt.hostServerPort, tt.hostServerDebugPort, tt.serverPort)

			if portMap == nil {
				t.Fatal("CreatePortBindings() returned nil")
			}

			// Check server port binding
			serverPortKey := nat.Port(strconv.Itoa(tt.serverPort))
			serverBindings, exists := (*portMap)[serverPortKey]
			if !exists {
				t.Errorf("Server port binding for %d not found", tt.serverPort)
			} else {
				if len(serverBindings) == 0 {
					t.Error("Server port bindings is empty")
				} else {
					binding := serverBindings[0]
					if binding.HostIP != constant.HostIP {
						t.Errorf("Server binding HostIP = %s, want %s", binding.HostIP, constant.HostIP)
					}
					if binding.HostPort != strconv.Itoa(tt.hostServerPort) {
						t.Errorf("Server binding HostPort = %s, want %s", binding.HostPort, strconv.Itoa(tt.hostServerPort))
					}
				}
			}

			// Check debug port binding
			debugPortKey := nat.Port(constant.DebugPort)
			debugBindings, exists := (*portMap)[debugPortKey]
			if !exists {
				t.Errorf("Debug port binding for %s not found", constant.DebugPort)
			} else {
				if len(debugBindings) == 0 {
					t.Error("Debug port bindings is empty")
				} else {
					binding := debugBindings[0]
					if binding.HostIP != constant.HostIP {
						t.Errorf("Debug binding HostIP = %s, want %s", binding.HostIP, constant.HostIP)
					}
					if binding.HostPort != strconv.Itoa(tt.hostServerDebugPort) {
						t.Errorf("Debug binding HostPort = %s, want %s", binding.HostPort, strconv.Itoa(tt.hostServerDebugPort))
					}
				}
			}
		})
	}
}

func TestCreateResources(t *testing.T) {
	tests := []struct {
		name      string
		isModule  bool
		resources map[string]any
		expectNil bool
	}{
		{
			name:      "default module resources",
			isModule:  true,
			resources: map[string]any{},
			expectNil: false,
		},
		{
			name:      "default sidecar resources",
			isModule:  false,
			resources: map[string]any{},
			expectNil: false,
		},
		{
			name:     "custom module resources",
			isModule: true,
			resources: map[string]any{
				"cpu_count":          2,
				"memory_reservation": 256,
				"memory":             1024,
				"memory_swap":        2048,
				"oom_kill_disable":   true,
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := CreateResources(tt.isModule, tt.resources)

			if tt.expectNil {
				if resources != nil {
					t.Error("Expected nil resources")
				}
				return
			}

			if resources == nil {
				t.Fatal("CreateResources() returned nil")
			}

			// Check that resources are properly configured
			if tt.isModule && len(tt.resources) == 0 {
				// Check default module values
				if resources.CPUCount != constant.ModuleCPU {
					t.Errorf("CPUCount = %d, want %d", resources.CPUCount, constant.ModuleCPU)
				}
				expectedMemory := ConvertMiBToBytes(constant.ModuleMemory)
				if resources.Memory != expectedMemory {
					t.Errorf("Memory = %d, want %d", resources.Memory, expectedMemory)
				}
			} else if !tt.isModule && len(tt.resources) == 0 {
				// Check default sidecar values
				if resources.CPUCount != constant.SidecarCPU {
					t.Errorf("CPUCount = %d, want %d", resources.CPUCount, constant.SidecarCPU)
				}
				expectedMemory := ConvertMiBToBytes(constant.SidecarMemory)
				if resources.Memory != expectedMemory {
					t.Errorf("Memory = %d, want %d", resources.Memory, expectedMemory)
				}
			}
		})
	}
}

func TestAppendAdditionalRequiredContainers(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name                   string
		initialContainers      []string
		moduleConfig           map[string]interface{}
		expectedContainers     []string
		expectedContainsSearch bool
		expectedContainsExport bool
	}{
		{
			name:                   "no modules enabled",
			initialContainers:      []string{"postgres"},
			moduleConfig:           map[string]interface{}{},
			expectedContainers:     []string{"postgres"},
			expectedContainsSearch: false,
			expectedContainsExport: false,
		},
		{
			name:              "search module enabled",
			initialContainers: []string{"postgres"},
			moduleConfig: map[string]interface{}{
				constant.ModSearchModule: map[string]interface{}{
					"deploy-module": true,
				},
			},
			expectedContainsSearch: true,
			expectedContainsExport: false,
		},
		{
			name:              "data export worker module enabled",
			initialContainers: []string{"postgres"},
			moduleConfig: map[string]interface{}{
				constant.ModDataExportWorkerModule: map[string]interface{}{
					"deploy-module": true,
				},
			},
			expectedContainsSearch: false,
			expectedContainsExport: true,
		},
		{
			name:              "both modules enabled",
			initialContainers: []string{"postgres"},
			moduleConfig: map[string]interface{}{
				constant.ModSearchModule: map[string]interface{}{
					"deploy-module": true,
				},
				constant.ModDataExportWorkerModule: map[string]interface{}{
					"deploy-module": true,
				},
			},
			expectedContainsSearch: true,
			expectedContainsExport: true,
		},
		{
			name:              "modules exist but disabled",
			initialContainers: []string{"postgres"},
			moduleConfig: map[string]interface{}{
				constant.ModSearchModule: map[string]interface{}{
					"deploy-module": false,
				},
				constant.ModDataExportWorkerModule: map[string]interface{}{
					"deploy-module": false,
				},
			},
			expectedContainsSearch: false,
			expectedContainsExport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set up test configuration
			viper.Reset()
			viper.Set("backend-modules", tt.moduleConfig)

			action := &action.Action{Name: "test"}
			result := AppendAdditionalRequiredContainers(action, tt.initialContainers)

			// Check that initial containers are preserved
			for _, container := range tt.initialContainers {
				found := false
				for _, resultContainer := range result {
					if resultContainer == container {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Initial container %s not found in result", container)
				}
			}

			// Check search module containers
			containsElasticsearch := false
			for _, container := range result {
				if container == constant.ElasticsearchContainer {
					containsElasticsearch = true
					break
				}
			}
			if containsElasticsearch != tt.expectedContainsSearch {
				t.Errorf("Expected Elasticsearch container: %v, got: %v", tt.expectedContainsSearch, containsElasticsearch)
			}

			// Check export worker module containers
			exportContainers := []string{constant.MinIOContainer, constant.CreateBucketsContainer, constant.FTPServerContainer}
			containsExportContainers := 0
			for _, container := range result {
				for _, exportContainer := range exportContainers {
					if container == exportContainer {
						containsExportContainers++
						break
					}
				}
			}
			expectedExportCount := 0
			if tt.expectedContainsExport {
				expectedExportCount = len(exportContainers)
			}
			if containsExportContainers != expectedExportCount {
				t.Errorf("Expected %d export containers, got %d", expectedExportCount, containsExportContainers)
			}
		})
	}
}

func TestIsModuleEnabled(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name         string
		module       string
		moduleConfig map[string]interface{}
		expected     bool
	}{
		{
			name:         "module not exists",
			module:       "non-existent-module",
			moduleConfig: map[string]interface{}{},
			expected:     false,
		},
		{
			name:   "module exists with deploy-module true",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": map[string]interface{}{
					"deploy-module": true,
				},
			},
			expected: true,
		},
		{
			name:   "module exists with deploy-module false",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": map[string]interface{}{
					"deploy-module": false,
				},
			},
			expected: false,
		},
		{
			name:   "module exists without deploy-module entry (legacy behavior)",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": map[string]interface{}{
					"version": "1.0.0",
				},
			},
			expected: true, // Should default to enabled
		},
		{
			name:   "module exists but value is nil",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": nil,
			},
			expected: false,
		},
		{
			name:   "module exists but value is not a map",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": "invalid-value",
			},
			expected: false,
		},
		{
			name:   "module exists with deploy-module as non-bool",
			module: "test-module",
			moduleConfig: map[string]interface{}{
				"test-module": map[string]interface{}{
					"deploy-module": "true", // String instead of bool
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set up test configuration
			viper.Reset()
			viper.Set("backend-modules", tt.moduleConfig)

			result := IsModuleEnabled(tt.module)
			if result != tt.expected {
				t.Errorf("IsModuleEnabled(%q) = %v, want %v", tt.module, result, tt.expected)
			}
		})
	}
}

func TestIsUIEnabled(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name         string
		tenant       string
		tenantConfig map[string]interface{}
		expected     bool
	}{
		{
			name:         "tenant not exists",
			tenant:       "non-existent-tenant",
			tenantConfig: map[string]interface{}{},
			expected:     false,
		},
		{
			name:   "tenant exists with deploy-ui true",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"deploy-ui": true,
				},
			},
			expected: true,
		},
		{
			name:   "tenant exists with deploy-ui false",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"deploy-ui": false,
				},
			},
			expected: false,
		},
		{
			name:   "tenant exists without deploy-ui entry",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"name": "Test Tenant",
				},
			},
			expected: false,
		},
		{
			name:   "tenant exists but value is nil",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": nil,
			},
			expected: false,
		},
		{
			name:   "tenant exists but value is not a map",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": "invalid-value",
			},
			expected: false,
		},
		{
			name:   "tenant exists with deploy-ui as non-bool",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"deploy-ui": "true", // String instead of bool
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set up test configuration
			viper.Reset()
			viper.Set("tenants", tt.tenantConfig)

			result := IsUIEnabled(tt.tenant)
			if result != tt.expected {
				t.Errorf("IsUIEnabled(%q) = %v, want %v", tt.tenant, result, tt.expected)
			}
		})
	}
}
