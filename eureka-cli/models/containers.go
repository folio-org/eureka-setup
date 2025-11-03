package models

// Containers represents a collection of container configurations and their associated metadata
type Containers struct {
	VaultRootToken   string
	RegistryHostname map[string]string
	RegistryModules  map[string][]*RegistryModule
	BackendModules   map[string]BackendModule
	GlobalEnv        []string
	SidecarEnv       []string
	ManagementOnly   bool
}

// NewCoreAndBusinessContainers creates a new Containers instance for core and business modules including sidecars
func NewCoreAndBusinessContainers(vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModules map[string]BackendModule,
	globalEnv []string,
	sidecarEnv []string) *Containers {
	return &Containers{
		VaultRootToken:   vaultRootToken,
		RegistryHostname: registryHosts,
		RegistryModules:  registryModules,
		BackendModules:   backendModules,
		GlobalEnv:        globalEnv,
		SidecarEnv:       sidecarEnv,
		ManagementOnly:   false,
	}
}

// NewManagementContainers creates a new Containers instance for management modules only (no sidecars)
func NewManagementContainers(vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModules map[string]BackendModule,
	globalEnv []string) *Containers {
	return &Containers{
		VaultRootToken:   vaultRootToken,
		RegistryHostname: registryHosts,
		RegistryModules:  registryModules,
		BackendModules:   backendModules,
		GlobalEnv:        globalEnv,
		SidecarEnv:       nil,
		ManagementOnly:   true,
	}
}
