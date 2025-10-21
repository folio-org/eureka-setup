package models

type RegistryModule struct {
	ID          string  `json:"id"`
	Action      string  `json:"action"`
	Name        string  `json:"-"`
	SidecarName string  `json:"-"`
	Version     *string `json:"-"`
}

type RegistryModules []RegistryModule

type RegistryModuleExtract struct {
	RegistryURLs         map[string]string
	RegistryModules      map[string][]*RegistryModule
	BackendModulesMap    map[string]BackendModule
	FrontendModulesMap   map[string]FrontendModule
	ModuleDescriptorsMap map[string]any
}

func NewRegistryModuleExtract(
	registryURLs map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	frontendModulesMap map[string]FrontendModule,
) *RegistryModuleExtract {

	return &RegistryModuleExtract{
		RegistryURLs:         registryURLs,
		RegistryModules:      registryModules,
		BackendModulesMap:    backendModulesMap,
		FrontendModulesMap:   frontendModulesMap,
		ModuleDescriptorsMap: make(map[string]any),
	}
}
