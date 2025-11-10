package models

// Containers represents a collection of container configurations and their associated metadata
type Containers struct {
	VaultRootToken string
	Modules        *ProxyModulesByRegistry
	BackendModules map[string]BackendModule
	GlobalEnv      []string
	SidecarEnv     []string
	ManagementOnly bool
}

// NewCoreAndBusinessContainers creates a new Containers instance for core and business modules including sidecars
func NewCoreAndBusinessContainers(vaultRootToken string,
	modules *ProxyModulesByRegistry,
	backendModules map[string]BackendModule,
	globalEnv []string,
	sidecarEnv []string) *Containers {
	return &Containers{
		VaultRootToken: vaultRootToken,
		Modules:        modules,
		BackendModules: backendModules,
		GlobalEnv:      globalEnv,
		SidecarEnv:     sidecarEnv,
		ManagementOnly: false,
	}
}

// NewManagementContainers creates a new Containers instance for management modules only (no sidecars)
func NewManagementContainers(vaultRootToken string,
	modules *ProxyModulesByRegistry,
	backendModules map[string]BackendModule,
	globalEnv []string) *Containers {
	return &Containers{
		VaultRootToken: vaultRootToken,
		Modules:        modules,
		BackendModules: backendModules,
		GlobalEnv:      globalEnv,
		SidecarEnv:     nil,
		ManagementOnly: true,
	}
}
