package models

type Containers struct {
	VaultRootToken    string
	RegistryHostname  map[string]string
	RegistryModules   map[string][]*RegistryModule
	BackendModulesMap map[string]BackendModule
	GlobalEnv         []string
	SidecarEnv        []string
	ManagementOnly    bool
}

func NewCoreAndBusinessContainers(vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnv []string,
	sidecarEnv []string) *Containers {
	return &Containers{
		VaultRootToken:    vaultRootToken,
		RegistryHostname:  registryHosts,
		RegistryModules:   registryModules,
		BackendModulesMap: backendModulesMap,
		GlobalEnv:         globalEnv,
		SidecarEnv:        sidecarEnv,
		ManagementOnly:    false,
	}
}

func NewManagementContainers(vaultRootToken string,
	registryHosts map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnv []string) *Containers {
	return &Containers{
		VaultRootToken:    vaultRootToken,
		RegistryHostname:  registryHosts,
		RegistryModules:   registryModules,
		BackendModulesMap: backendModulesMap,
		GlobalEnv:         globalEnv,
		SidecarEnv:        nil,
		ManagementOnly:    true,
	}
}
