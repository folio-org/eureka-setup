package internal

const (
	ProfileKey     string = "profile"
	ProfileNameKey string = "profile.name"

	ApplicationKey                              string = "application"
	ApplicationNameKey                          string = "application.name"
	ApplicationVersionKey                       string = "application.version"
	ApplicationPlatformKey                      string = "application.platform"
	ApplicationFetchDescriptorsKey              string = "application.fetch-descriptors"
	ApplicationPortStartKey                     string = "application.port-start"
	ApplicationPortEndKey                       string = "application.port-end"
	ApplicationPlatformCompleteStripesBranchKey string = "application.stripes-branch"
	ApplicationGatewayHostnameKey               string = "application.gateway-hostname"
	ApplicationGatewayDependenciesKey           string = "application.dependencies"

	RegistryKey    string = "registry"
	RegistryUrlKey string = "registry.url"

	InstallKey       string = "install"
	InstallFolioKey  string = "install.folio"
	InstallEurekaKey string = "install.eureka"

	NamespacesKey                   string = "namespaces"
	NamespacesPlatformCompleteUiKey string = "namespaces.platform-complete-ui"

	EnvironmentKey      string = "environment"
	EnvironmentFolioKey string = "environment.ENV"

	TenantsKey                 string = "tenants"
	TenantsDeployUiEntryKey    string = "deploy-ui"
	TenantsSingleTenantKey     string = "single-tenant"
	TenantsEnableEcsRequestKey string = "enable-ecs-request"

	UsersKey string = "users"
	RolesKey string = "roles"

	SidecarModuleKey                string = "sidecar-module"
	SidecarModuleEnvironmentKey     string = "sidecar-module.environment"
	SidecarModuleResourcesKey       string = "sidecar-module.resources"
	SidecarModuleImageEntryKey      string = "image"
	SidecarModuleLocalImageEntryKey string = "local-image"
	SidecarModuleVersionEntryKey    string = "version"

	BackendModulesKey        string = "backend-modules"
	FrontendModulesKey       string = "frontend-modules"
	CustomFrontendModulesKey string = "custom-frontend-modules"

	ModuleDeployModuleEntryKey              string = "deploy-module"
	ModuleDeploySidecarEntryKey             string = "deploy-sidecar"
	ModuleVersionEntryKey                   string = "version"
	ModulePortEntryKey                      string = "port"
	ModulePortServerEntryKey                string = "port-server"
	ModuleUseVaultEntryKey                  string = "use-vault"
	ModuleUseOkapiUrlEntryKey               string = "use-okapi-url"
	ModuleDisableSystemUserEntryKey         string = "disable-system-user"
	ModuleEnvironmentEntryKey               string = "environment"
	ModuleVolumesEntryKey                   string = "volumes"
	ModuleResourceEntryKey                  string = "resources"
	ModuleResourceCpuCountEntryKey          string = "cpu-count"
	ModuleResourceMemoryReservationEntryKey string = "memory-reservation"
	ModuleResourceMemoryEntryKey            string = "memory"
	ModuleResourceMemorySwapEntryKey        string = "memory-swap"
	ModuleResourceOomKillDisableEntryKey    string = "oom-kill-disable"
)
