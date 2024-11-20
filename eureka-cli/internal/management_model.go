package internal

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name        string
	SidecarName string
	Version     *string
}

type RegistryModules []RegistryModule
