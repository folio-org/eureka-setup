package models

// FrontendModule represents configuration for a frontend module
type FrontendModule struct {
	DeployModule        bool
	ModuleVersion       *string
	ModuleName          string
	LocalDescriptorPath string
}

// NewFrontendModule creates a new FrontendModule instance
func NewFrontendModule(deployModule bool, name string, version *string, localDescriptorPath string) *FrontendModule {
	return &FrontendModule{DeployModule: deployModule, ModuleName: name, ModuleVersion: version, LocalDescriptorPath: localDescriptorPath}
}
