package internal

const (
	ProfileNameKey string = "profile.name"

	ApplicationKey                string = "application"
	ApplicationPortStart          string = "application.port-start"
	ApplicationStripesBranchKey   string = "application.stripes-branch"
	ApplicationGatewayHostnameKey string = "application.gateway-hostname"

	RegistryUrlKey                          string = "registry.registry-url"
	RegistryFolioInstallJsonUrlKey          string = "registry.folio-install-json-url"
	RegistryEurekaInstallJsonUrlKey         string = "registry.eureka-install-json-url"
	RegistryNamespacesPlatformCompleteUiKey string = "registry.namespaces.platform-complete-ui"

	EnvironmentKey string = "environment"

	TenantsKey string = "tenants"
	UsersKey   string = "users"
	RolesKey   string = "roles"

	SidecarModule               string = "sidecar-module"
	SidecarModuleEnvironmentKey string = "sidecar-module.environment"
	BackendModuleKey            string = "backend-modules"
	FrontendModuleKey           string = "frontend-modules"
	CustomFrontendModuleKey     string = "custom-frontend-modules"
)
