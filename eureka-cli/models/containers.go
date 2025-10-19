package models

type Containers struct {
	VaultRootToken     string
	RegistryHostname   map[string]string
	RegistryModules    map[string][]*RegistryModule
	BackendModulesMap  map[string]BackendModule
	GlobalEnvironment  []string
	SidecarEnvironment []string
	ManagementOnly     bool
}

func NewCoreAndBusinessContainers(
	vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string,
	sidecarEnvironment []string,
) *Containers {

	return &Containers{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHosts,
		RegistryModules:    registryModules,
		BackendModulesMap:  backendModulesMap,
		GlobalEnvironment:  globalEnvironment,
		SidecarEnvironment: sidecarEnvironment,
		ManagementOnly:     false,
	}
}

func NewManagementContainers(
	vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string,
) *Containers {

	return &Containers{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHosts,
		RegistryModules:    registryModules,
		BackendModulesMap:  backendModulesMap,
		GlobalEnvironment:  globalEnvironment,
		SidecarEnvironment: nil,
		ManagementOnly:     true,
	}
}
