package models

type FrontendModule struct {
	DeployModule  bool
	ModuleVersion *string
	ModuleName    string
}

func NewFrontendModule(deployModule bool, name string, version *string) *FrontendModule {
	return &FrontendModule{
		DeployModule:  true,
		ModuleName:    name,
		ModuleVersion: version,
	}
}
