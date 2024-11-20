package internal

const (
	ProfileNameKey string = "profile.name"

	ApplicationKey           string = "application"
	ApplicationPortStart     string = "application.port-start"
	ApplicationStripesBranch string = "application.stripes-branch"

	RegistryUrlKey                  string = "registry.registry-url"
	RegistryFolioInstallJsonUrlKey  string = "registry.folio-install-json-url"
	RegistryEurekaInstallJsonUrlKey string = "registry.eureka-install-json-url"

	EnvironmentKey string = "environment"

	TenantsKey string = "tenants"
	UsersKey   string = "users"
	RolesKey   string = "roles"

	SidecarModule               string = "sidecar-module"
	SidecarModuleEnvironmentKey string = "sidecar-module.environment"
	BackendModuleKey            string = "backend-modules"
	FrontendModuleKey           string = "frontend-modules"
)
