package internal

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name        string
	SidecarName string
	Version     *string
}

type RegistryModules []RegistryModule

type Applications struct {
	ApplicationDescriptors []map[string]any `json:"applicationDescriptors"`
	TotalRecords           int              `json:"totalRecords"`
}

type ConsortiumTenant struct {
	Tenant    string
	IsCentral int
}

type ConsortiumTenants []*ConsortiumTenant
