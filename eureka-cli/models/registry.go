package models

// ProxyModulesResponse represents the response containing a list of proxy modules from the registry
type ProxyModulesResponse []ProxyModule

// ProxyModule represents a proxy module with ID and metadata
type ProxyModule struct {
	ID     string `json:"id"`
	Action string `json:"action,omitempty"`
	ProxyModuleMetadata
}

// ProxyModuleMetadata represents proxy module metadata
type ProxyModuleMetadata struct {
	Name        string  `json:"-"`
	SidecarName string  `json:"-"`
	Version     *string `json:"-"`
}

// ProxyModulesByRegistry organizes proxy modules by their registry source
type ProxyModulesByRegistry struct {
	FolioModules  []*ProxyModule
	EurekaModules []*ProxyModule
}

// NewProxyModulesByRegistry creates a new ProxyModulesByRegistry instance
func NewProxyModulesByRegistry(folioModules, eurekaModules []*ProxyModule) *ProxyModulesByRegistry {
	return &ProxyModulesByRegistry{
		FolioModules:  folioModules,
		EurekaModules: eurekaModules,
	}
}

// RegistryExtract contains extracted information about modules from registries
type RegistryExtract struct {
	RegistryURLs      map[string]string
	Modules           *ProxyModulesByRegistry
	BackendModules    map[string]BackendModule
	FrontendModules   map[string]FrontendModule
	ModuleDescriptors map[string]any
}

// NewRegistryExtract creates a new RegistryModuleExtract instance
func NewRegistryExtract(registryURLs map[string]string,
	modules *ProxyModulesByRegistry,
	backendModules map[string]BackendModule,
	frontendModules map[string]FrontendModule) *RegistryExtract {
	return &RegistryExtract{
		RegistryURLs:      registryURLs,
		Modules:           modules,
		BackendModules:    backendModules,
		FrontendModules:   frontendModules,
		ModuleDescriptors: make(map[string]any),
	}
}
