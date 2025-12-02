package action

// Param is a central container of all parameters
// passed to the program by the user from the shell instance
type Param struct {
	All                 bool
	ApplicationNames    []string
	BuildImages         bool
	ConfigFile          string
	DefaultGateway      bool
	EnableDebug         bool
	EnableECSRequests   bool
	GatewayHostname     string
	GatewayURL          string
	ID                  string
	Length              int
	ModuleName          string
	ModuleType          string
	ModuleURL           string
	Namespace           string
	OnlyRequired        bool
	OverwriteFiles      bool
	PlatformCompleteURL string
	PrivatePort         int
	Profile             string
	PurgeSchemas        bool
	RemoveApplication   bool
	Restore             bool
	SidecarURL          string
	SingleTenant        bool
	SkipApplication     bool
	SkipCapabilitySets  bool
	SkipRegistry        bool
	Tenant              string
	TenantIDs           []string
	TokenType           string
	UpdateCloned        bool
	User                string
	Versions            int
}

// Flag holds the metadata for a CLI flag
type Flag struct {
	Long        string
	Short       string
	Description string
}

// GetName returns the flag name
func (f Flag) GetName() string {
	return f.Long
}

// Flag definitions
var (
	All                 = Flag{"all", "a", "All modules for all profiles"}
	ApplicationNames    = Flag{"apps", "", "Application names"}
	BuildImages         = Flag{"buildImages", "b", "Build Docker images"}
	ConfigFile          = Flag{"configFile", "c", "Use a specific config file"}
	DefaultGateway      = Flag{"defaultGateway", "g", "Use default gateway in URLs, .e.g. http://host.docker.internal:{{port}} will be set automatically"}
	EnableDebug         = Flag{"enableDebug", "d", "Enable debug"}
	EnableECSRequests   = Flag{"enableEcsRequests", "", "Enable ECS requests"}
	GatewayHostname     = Flag{"gatewayHostname", "", "Gateway hostname"}
	GatewayURL          = Flag{"gatewayURL", "", "Gateway URL"}
	ID                  = Flag{"id", "i", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021"}
	Length              = Flag{"length", "l", "Salt length"}
	ModuleName          = Flag{"moduleName", "n", "Module name, e.g. mod-orders"}
	ModuleType          = Flag{"moduleType", "y", "Module type, e.g. management"}
	ModuleURL           = Flag{"moduleUrl", "m", "Module URL, e.g. http://host.docker.internal:36002 or 36002 (if -g is used)"}
	Namespace           = Flag{"namespace", "", "DockerHub namespace"}
	OnlyRequired        = Flag{"onlyRequired", "q", "Use only required system containers"}
	OverwriteFiles      = Flag{"overwriteFiles", "o", "Overwrite files in %s home directory"}
	PlatformCompleteURL = Flag{"platformCompleteURL", "", "Platform Complete UI url"}
	PrivatePort         = Flag{"privatePort", "", "Private port e.g. 8081"}
	Profile             = Flag{"profile", "p", "Use a specific profile, options: %s"}
	PurgeSchemas        = Flag{"purgeSchemas", "", "Purge schemas in PostgreSQL on uninstallation"}
	RemoveApplication   = Flag{"removeApplication", "", "Remove application from the DB"}
	Restore             = Flag{"restore", "r", "Restore module & sidecar"}
	SidecarURL          = Flag{"sidecarUrl", "s", "Sidecar URL e.g. http://host.docker.internal:37002 or 37002 (if -g is used)"}
	SingleTenant        = Flag{"singleTenant", "", "Use for Single Tenant workflow"}
	SkipCapabilitySets  = Flag{"skipCapabilitySets", "", "Skip refreshing capability sets"}
	SkipRegistry        = Flag{"skipRegistry", "", "Skip retrieving module registry versions"}
	Tenant              = Flag{"tenant", "t", "Tenant"}
	TokenType           = Flag{"tokenType", "", "Token type"}
	TenantIDs           = Flag{"ids", "", "Tenant ids"}
	UpdateCloned        = Flag{"updateCloned", "u", "Update Git cloned projects"}
	User                = Flag{"user", "x", "User"}
	Versions            = Flag{"versions", "v", "Number of versions, e.g. 5"}
)
