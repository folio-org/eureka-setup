package actionparams

// ActionParams is a central container of all parameters
// passed to the program by the user from the shell instance
type ActionParams struct {
	ConfigFile          string
	Profile             string
	OverwriteFiles      bool
	ModuleName          string
	EnableDebug         bool
	BuildImages         bool
	UpdateCloned        bool
	SingleTenant        bool
	EnableECSRequests   bool
	Tenant              string
	Namespace           string
	PlatformCompleteURL string
	All                 bool
	ID                  string
	ModuleURL           string
	SidecarURL          string
	Restore             bool
	DefaultGateway      bool
	OnlyRequired        bool
	User                string
	Length              int
	ModuleType          string
	PurgeSchemas        bool
	Lines               int
}
