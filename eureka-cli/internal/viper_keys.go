package internal

const (
	ProfileKey     string = "profile"
	ProfileNameKey string = "profile.name"

	ApplicationKey                    string = "application"
	ApplicationPortStartKey           string = "application.port-start"
	ApplicationPortEndKey             string = "application.port-end"
	ApplicationStripesBranchKey       string = "application.stripes-branch"
	ApplicationGatewayHostnameKey     string = "application.gateway-hostname"
	ApplicationGatewayDependenciesKey string = "application.dependencies"

	RegistryKey    string = "registry"
	RegistryUrlKey string = "registry.url"

	InstallKey       string = "install"
	InstallFolioKey  string = "install.folio"
	InstallEurekaKey string = "install.eureka"

	NamespacesKey                   string = "namespaces"
	NamespacesPlatformCompleteUiKey string = "namespaces.platform-complete-ui"

	EnvironmentKey      string = "environment"
	EnvironmentFolioKey string = "environment.ENV"

	TenantsKey              string = "tenants"
	TenantsDeployUiEntryKey string = "deploy-ui"

	UsersKey string = "users"
	RolesKey string = "roles"

	SidecarModuleKey             string = "sidecar-module"
	SidecarModuleEnvironmentKey  string = "sidecar-module.environment"
	SidecarModuleResourcesKey    string = "sidecar-module.resources"
	SidecarModuleImageEntryKey   string = "image"
	SidecarModuleVersionEntryKey string = "version"

	BackendModuleKey                                string = "backend-modules"
	BackendModuleModSearchKey                       string = "backend-modules.mod-search"
	BackendModuleModSearchDeployModuleKey           string = "backend-modules.mod-search.deploy-module"
	BackendModuleModDataExportWorkerKey             string = "backend-modules.mod-data-export-worker"
	BackendModuleModDataExportWorkerDeployModuleKey string = "backend-modules.mod-data-export-worker.deploy-module"

	FrontendModuleKey       string = "frontend-modules"
	CustomFrontendModuleKey string = "custom-frontend-modules"

	ModuleDeployModuleEntryKey              string = "deploy-module"
	ModuleDeploySidecarEntryKey             string = "deploy-sidecar"
	ModuleVersionEntryKey                   string = "version"
	ModulePortEntryKey                      string = "port"
	ModulePortServerEntryKey                string = "port-server"
	ModuleEnvironmentEntryKey               string = "environment"
	ModuleResourceEntryKey                  string = "resources"
	ModuleResourceCpuCountEntryKey          string = "cpu-count"
	ModuleResourceMemoryReservationEntryKey string = "memory-reservation"
	ModuleResourceMemoryEntryKey            string = "memory"
	ModuleResourceMemorySwapEntryKey        string = "memory-swap"
	ModuleResourceOomKillDisableEntryKey    string = "oom-kill-disable"
	ModuleUseVaultEntryKey                  string = "use-vault"
	ModuleDisableSystemUserEntryKey         string = "disable-system-user"
)
