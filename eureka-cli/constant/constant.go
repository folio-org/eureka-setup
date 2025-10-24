package constant

import (
	"time"
)

const (
	// Keycloak credentials
	KeycloakAdminUsername = "admin"
	KeycloakAdminPassword = "admin"

	// Command wait durations
	DeployApplicationPartitionWait    = 10 * time.Second
	DeploySystemWait                  = 15 * time.Second
	DeployAdditionalSystemWait        = 15 * time.Second
	DeployManagementWait              = 10 * time.Second
	DeployModulesWait                 = 10 * time.Second
	ModuleReadinessCheckWait          = 10 * time.Second
	AttachCapabilitySetsPollWait      = 30 * time.Second
	AttachCapabilitySetsRebalanceWait = 30 * time.Second
	AttachCapabilitySetsTimeoutWait   = 30 * time.Second

	// Vault client properties
	VaultTimeout = 30 * time.Second

	// Default HTTP client properties
	HTTPClientTimeout                = 30 * time.Minute
	HTTPClientDialTimeout            = 5 * time.Minute
	HTTPClientKeepAlive              = 90 * time.Second
	HTTPClientMaxIdleConns           = 50
	HTTPClientMaxIdleConnsPerHost    = 10
	HTTPClientIdleConnTimeout        = 120 * time.Second
	HTTPClientMaxResponseHeaderBytes = 16 << 20
	HTTPClientWriteBufferSize        = 64 << 10
	HTTPClientReadBufferSize         = 64 << 10
	HTTPClientResponseHeaderTimeout  = 5 * time.Minute
	HTTPClientExpectContinueTimeout  = 10 * time.Second
	HTTPClientDisableCompression     = false
	HTTPClientForceAttemptHTTP2      = false

	// Retry HTTP client properties
	RetryHTTPClientRetryMax     = 10
	RetryHTTPClientRetryWaitMin = 3 * time.Second
	RetryHTTPClientRetryWaitMax = 10 * time.Second

	// Container types
	ManagementType = "management"
	ModuleType     = "module"
	SidecarType    = "sidecar"

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
	ModuleMemoryReservation = 128
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

	// Module health check retries
	ModuleReadinessMaxRetries = 50

	// Consortium properties
	NoneConsortium = "nop"

	// Config
	ConfigDir  = ".eureka"
	ConfigType = "yaml"

	// Module registries
	FolioRegistry  = "folio"
	EurekaRegistry = "eureka"

	// Docker compose properties
	DockerComposeWorkDir = "./misc"

	// Container network properties
	NetworkID       = "eureka"
	NetworkAlias    = "eureka-net"
	DockerHostname  = "host.docker.internal"
	DockerGatewayIP = "172.17.0.1"
	HostIP          = "0.0.0.0"
	ServerPort      = "8081"
	DebugPort       = "5005"

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
	ModuleIDPattern       = "([a-z-_]+)([\\d-_.]+)([a-zA-Z0-9-_.]+)"
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

	// System container ports
	KongPort        = "8000"
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
	CombinedProfile  = "combined"
	ExportProfile    = "export"
	SearchProfile    = "search"
	EdgeProfile      = "edge"
	ECSProfile       = "ecs"
	ECSSingleProfile = "ecs-single"
	ImportProfile    = "import"
)

type TenantType string

const (
	All     TenantType = ""
	Default TenantType = "default"
	Central TenantType = "central"
	Member  TenantType = "member"
)

func Get() []TenantType {
	return []TenantType{Central, Member}
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
