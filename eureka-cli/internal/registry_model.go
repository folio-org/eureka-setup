package internal

type RegisterModuleDto struct {
	RegistryUrls         map[string]string
	RegistryModules      map[string][]*RegistryModule
	BackendModulesMap    map[string]BackendModule
	FrontendModulesMap   map[string]FrontendModule
	ModuleDescriptorsMap map[string]interface{}
	EnableDebug          bool
}

func NewRegisterModuleDto(registryUrls map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	frontendModulesMap map[string]FrontendModule,
	enableDebug bool) *RegisterModuleDto {
	return &RegisterModuleDto{
		RegistryUrls:         registryUrls,
		RegistryModules:      registryModules,
		BackendModulesMap:    backendModulesMap,
		FrontendModulesMap:   frontendModulesMap,
		ModuleDescriptorsMap: make(map[string]interface{}),
		EnableDebug:          enableDebug,
	}
}
