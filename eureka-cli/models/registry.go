package models

// RegistryModule represents a module fetched from a registry
type RegistryModule struct {
	ID          string  `json:"id"`
	Action      string  `json:"action"`
	Name        string  `json:"-"`
	SidecarName string  `json:"-"`
	Version     *string `json:"-"`
}

// RegistryModules represents a collection of registry modules
type RegistryModules []RegistryModule

// RegistryModuleExtract contains extracted information about modules from registries
type RegistryModuleExtract struct {
	RegistryURLs      map[string]string
	RegistryModules   map[string][]*RegistryModule
	BackendModules    map[string]BackendModule
	FrontendModules   map[string]FrontendModule
	ModuleDescriptors map[string]any
}

// NewRegistryModuleExtract creates a new RegistryModuleExtract instance
func NewRegistryModuleExtract(registryURLs map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModules map[string]BackendModule,
	frontendModules map[string]FrontendModule) *RegistryModuleExtract {
	return &RegistryModuleExtract{
		RegistryURLs:      registryURLs,
		RegistryModules:   registryModules,
		BackendModules:    backendModules,
		FrontendModules:   frontendModules,
		ModuleDescriptors: make(map[string]any),
	}
}
