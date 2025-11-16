package constant

import "time"

const (
	// Command wait durations
	DeployApplicationPartitionWait    = 15 * time.Second
	DeploySystemWait                  = 15 * time.Second
	DeployAdditionalSystemWait        = 15 * time.Second
	DeployManagementWait              = 5 * time.Second
	DeployModulesWait                 = 5 * time.Second
	ModuleReadinessWait               = 10 * time.Second
	KongReadinessWait                 = 10 * time.Second
	AttachCapabilitySetsPollWait      = 30 * time.Second
	AttachCapabilitySetsRebalanceWait = 30 * time.Second
	AttachCapabilitySetsTimeoutWait   = 30 * time.Second
	ConsortiumTenantStatusWait        = 10 * time.Second

	// Readiness retries
	ModuleReadinessMaxRetries     = 50
	KongRouteReadinessMaxRetries  = 30
	ConsumerGroupRebalanceRetries = 30
	ConsumerGroupPollMaxRetries   = 30

	// Context timeout durations
	ContextTimeoutDockerAPIVersion   = 15 * time.Second
	ContextTimeoutDockerList         = 30 * time.Second
	ContextTimeoutDockerImagePull    = 5 * time.Minute
	ContextTimeoutDockerDeploy       = 2 * time.Minute
	ContextTimeoutDockerUndeploy     = 1 * time.Minute
	ContextTimeoutVaultClient        = 30 * time.Second
	ContextTimeoutVaultContainerLogs = 30 * time.Second
	ContextTimeoutAWSConfig          = 30 * time.Second

	// HTTP client timeouts
	HTTPClientPingTimeout = 15 * time.Second
	HTTPClientTimeout     = 10 * time.Minute

	// Custom HTTP client transport settings
	HTTPClientDialTimeout           = 30 * time.Second
	HTTPClientKeepAlive             = 30 * time.Second
	HTTPClientMaxIdleConns          = 10
	HTTPClientMaxIdleConnsPerHost   = 2
	HTTPClientIdleConnTimeout       = 90 * time.Second
	HTTPClientResponseHeaderTimeout = 5 * time.Minute
	HTTPClientExpectContinueTimeout = 1 * time.Second

	// Ping HTTP client transport settings
	HTTPClientPingDialTimeout           = 5 * time.Second
	HTTPClientPingKeepAlive             = -1 // Disable TCP keep-alive
	HTTPClientPingDisableKeepAlives     = true
	HTTPClientPingMaxIdleConns          = 0
	HTTPClientPingMaxIdleConnsPerHost   = 0
	HTTPClientPingIdleConnTimeout       = 0
	HTTPClientPingResponseHeaderTimeout = 5 * time.Second

	// Docker log properties
	DockerLogHeaderSize = 8
	DockerLogSizeOffset = 4

	// Retry HTTP client properties
	RetryHTTPClientRetryMax     = 5
	RetryHTTPClientRetryWaitMin = 2 * time.Second
	RetryHTTPClientRetryWaitMax = 10 * time.Second

	SidecarProjectName = "folio-module-sidecar"

	// Folio source Git repository URLs
	FolioKongRepositoryURL        = "https://github.com/folio-org/folio-kong"
	FolioKeycloakRepositoryURL    = "https://github.com/folio-org/folio-keycloak"
	PlatformCompleteRepositoryURL = "https://github.com/folio-org/platform-complete.git"

	// Folio source Git repository labels
	FolioKongLabel        = "folio-kong"
	FolioKeycloakLabel    = "folio-keycloak"
	PlatformCompleteLabel = "platform-complete"

	// Folio source Git local repository output directories
	FolioKongOutputDir        = "folio-kong"
	FolioKeycloakOutputDir    = "folio-keycloak"
	PlatformCompleteOutputDir = "platform-complete"

	// Branch names
	FolioKongBranch     = "master"
	FolioKeycloakBranch = "master"
	StripesBranch       = "snapshot"

	// Docker Hub registries
	SnapshotRegistry = "folioci"
	ReleaseRegistry  = "folioorg"

	// AWS ECR env var name
	ECRRepositoryEnv = "AWS_ECR_FOLIO_REPO"

	// Container resources
	ModuleCPU               = 1
	ModuleMemoryReservation = 120
	ModuleMemory            = 750
	ModuleSwap              = -1

	// Sidecar resources
	SidecarCPU               = 1
	SidecarMemoryReservation = 64
	SidecarMemory            = 450
	SidecarSwap              = -1

	// Charset for key generation
	Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// HTTP Headers
	ApplicationJSON           = "application/json"
	ApplicationFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeHeader         = "Content-Type"
	AuthorizationHeader       = "Authorization"
	OkapiTenantHeader         = "X-Okapi-Tenant"
	OkapiTokenHeader          = "X-Okapi-Token"

	// Consortium properties
	NoneConsortium = "nop"

	// Config
	ConfigPrefix = "config"
	ConfigDir    = ".eureka"
	ConfigType   = "yaml"

	// Module registries
	FolioRegistry  = "folio"
	EurekaRegistry = "eureka"

	// Docker compose properties
	DockerComposeWorkDir = "./misc"

	// Container network properties
	NetworkID         = "eureka"
	NetworkAlias      = "eureka-net"
	DockerHostname    = "host.docker.internal"
	DockerGatewayIP   = "172.17.0.1"
	HostIP            = "0.0.0.0"
	PrivateServerPort = "8081"
	PrivateDebugPort  = "5005"

	// Container regexp patterns
	ManagementModulePattern               = "mgr-"
	EdgeModulePattern                     = "edge-"
	AllContainerPattern                   = "^eureka-"
	ProfileContainerPattern               = "^eureka-%s"
	ManagementContainerPattern            = "^eureka-mgr-"
	ModuleContainerPattern                = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?$"
	SidecarContainerPattern               = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?-sc$"
	SingleModuleOrSidecarContainerPattern = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
	SingleUiContainerPattern              = "eureka-platform-complete-ui-%s"

	// Other regexp patterns
	ColonDelimitedPattern = ".*:"
	ModuleIDPattern       = `^([a-z_-]+)([\d_.-]+)([-\w.]+)$`
	VaultRootTokenPattern = "init.sh: Root VAULT TOKEN is:"
	NewLinePattern        = `[\r\n\s-]+`

	// System containers name
	PostgreSQLContainer    = "postgres"
	KafkaContainer         = "kafka"
	KafkaToolsContainer    = "kafka-tools"
	KeycloakProxyContainer = "keycloak"
	KeycloakContainer      = "keycloak-internal"
	KongContainer          = "kong"
	VaultContainer         = "vault"
	ElasticsearchContainer = "elasticsearch"
	MinIOContainer         = "minio"
	CreateBucketsContainer = "createbuckets"
	FTPServerContainer     = "ftp-server"

	// Keycloak credentials
	KeycloakAdminUsername = "admin"
	KeycloakAdminPassword = "admin"

	// System container ports
	KongPort        = "8000"
	KongAdminPort   = "8001"
	VaultServerPort = "8200"

	// System container internal endpoints
	KafkaTCP     = "kafka.eureka:9092"
	VaultHTTP    = "http://vault.eureka:8200"
	KeycloakHTTP = "http://keycloak.eureka:8080"

	// System container external endpoints
	KongExternalHTTP     = "http://localhost:8000"
	KeycloakExternalHTTP = "http://keycloak.eureka:8080"

	// Backend modules
	ModSearchModule           = "mod-search"
	ModDataExportWorkerModule = "mod-data-export-worker"

	// Kafka consumer group properties
	ConsumerGroupSuffix = "mod-roles-keycloak-capability-group"
	ErrNoActiveMembers  = "Consumer group 'folio-mod-roles-keycloak-capability-group' has no active members."
	ErrRebalancing      = "Consumer group 'folio-mod-roles-keycloak-capability-group' is rebalancing."
	ErrTimeoutException = "TimeoutException"

	// Profile names
	CombinedProfile       = "combined"
	CombinedNativeProfile = "combined-native"
	ExportProfile         = "export"
	SearchProfile         = "search"
	EdgeProfile           = "edge"
	ECSProfile            = "ecs"
	ECSSingleProfile      = "ecs-single"
	ImportProfile         = "import"
)

// Container types
const (
	Management = "management"
	Module     = "module"
	Sidecar    = "sidecar"
)

func GetContainerTypes() []string {
	return []string{Module, Sidecar, Management}
}

// Tenant types
type TenantType string

const (
	All     = ""
	Default = "default"
	Central = "central"
	Member  = "member"
)

func GetTenantTypes() []TenantType {
	return []TenantType{Central, Member}
}

// Keycloak Grant types
type KeycloakGrantType string

const (
	ClientCredentials = "client_credentials"
	Password          = "password"
)

const (
	DefaultToken        string = "tenant"
	MasterCustomToken   string = "master-custom"
	MasterAdminCLIToken string = "master-admin-cli"
)

func GetTokenTypes() []string {
	return []string{DefaultToken, MasterCustomToken, MasterAdminCLIToken}
}

func GetInitialRequiredContainers() []string {
	return []string{
		PostgreSQLContainer,
		KafkaContainer,
		KafkaToolsContainer,
		VaultContainer,
		KeycloakProxyContainer,
		KeycloakContainer,
		KongContainer,
	}
}

func GetProfiles() []string {
	return []string{
		CombinedProfile,
		CombinedNativeProfile,
		ExportProfile,
		SearchProfile,
		EdgeProfile,
		ECSProfile,
		ECSSingleProfile,
		ImportProfile,
	}
}

func GetDefaultProfile() string {
	return CombinedProfile
}
