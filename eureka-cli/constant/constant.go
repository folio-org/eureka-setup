package constant

import (
	"time"
)

const (
	KafkaTCP     = "kafka.eureka:9092"
	VaultHTTP    = "http://vault.eureka:8200"
	KeycloakHTTP = "http://keycloak.eureka:8080"

	KongExternalHTTP     = "http://localhost:8000"
	KeycloakExternalHTTP = "http://keycloak.eureka:8080"

	KeycloakAdminUsername = "admin"
	KeycloakAdminPassword = "admin"

	VaultServerPort = 8200
	VaultTimeout    = 30 * time.Second

	DefaultHTTPRetryMax     = 10
	DefaultHTTPRetryWaitMin = 3 * time.Second
	DefaultHTTPRetryWaitMax = 10 * time.Second

	FolioKongRepositoryURL        = "https://github.com/folio-org/folio-kong"
	FolioKeycloakRepositoryURL    = "https://github.com/folio-org/folio-keycloak"
	PlatformCompleteRepositoryURL = "https://github.com/folio-org/platform-complete.git"

	Module     = "module"
	Sidecar    = "sidecar"
	Management = "management"

	SidecarProjectName               = "folio-module-sidecar"
	DefaultFolioKongOutputDir        = "folio-kong"
	DefaultFolioKeycloakOutputDir    = "folio-keycloak"
	DefaultPlatformCompleteOutputDir = "platform-complete"
	DefaultFolioKongBranch           = "master"
	DefaultFolioKeycloakBranch       = "master"
	DefaultStripesBranch             = "snapshot"
	SnapshotRegistry                 = "folioci"
	ReleaseRegistry                  = "folioorg"
	ECRRepository                    = "AWS_ECR_FOLIO_REPO"

	DefaultModuleCPU                = 1
	DefaultModuleMemoryReservation  = 128
	DefaultModuleMemory             = 750
	DefaultModuleSwap               = -1
	DefaultSidecarCPU               = 1
	DefaultSidecarMemoryReservation = 64
	DefaultSidecarMemory            = 450
	DefaultSidecarSwap              = -1

	Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	JsonContentType           = "application/json"
	FormURLEncodedContentType = "application/x-www-form-urlencoded"

	ContentTypeHeader   = "Content-Type"
	AuthorizationHeader = "Authorization"
	TenantHeader        = "X-Okapi-Tenant"
	TokenHeader         = "X-Okapi-Token"

	HealthCheckMaxAttempts = 50
	NoneConsortium         = "nop"

	ConfigDir              = ".eureka"
	ConfigType             = "yaml"
	FolioRegistry          = "folio"
	EurekaRegistry         = "eureka"
	DockerComposeWorkDir   = "./misc"
	DefaultNetworkID       = "eureka"
	DefaultNetworkAlias    = "eureka-net"
	DefaultDockerHostname  = "host.docker.internal"
	DefaultDockerGatewayIP = "172.17.0.1"
	DefaultHostIP          = "0.0.0.0"
	GatewayPort            = 8000
	DefaultServerPort      = "8081"
	DefaultDebugPort       = "5005"

	ColonDelimitedPattern                 = ".*:"
	ManagementModulePattern               = "mgr-"
	EdgeModulePattern                     = "edge-"
	AllContainerPattern                   = "^eureka-"
	ProfileContainerPattern               = "^eureka-%s"
	ManagementContainerPattern            = "^eureka-mgr-"
	ModuleContainerPattern                = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?$"
	SidecarContainerPattern               = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?-sc$"
	SingleModuleOrSidecarContainerPattern = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
	SingleUiContainerPattern              = "eureka-platform-complete-ui-%s"
	ModuleIDPattern                       = "([a-z-_]+)([\\d-_.]+)([a-zA-Z0-9-_.]+)"
	VaultRootTokenPattern                 = "init.sh: Root VAULT TOKEN is:"
	NewLinePattern                        = `[\r\n\s-]+`

	VaultContainerName            = "vault"
	ElasticsearchContainerName    = "elasticsearch"
	MinioContainerName            = "minio"
	CreateBucketsContainerName    = "createbuckets"
	FtpServerContainerName        = "ftp-server"
	ModSearchModuleName           = "mod-search"
	ModDataExportWorkerModuleName = "mod-data-export-worker"

	ConsumerGroupSuffix = "mod-roles-keycloak-capability-group"
	ErrNoActiveMembers  = "Consumer group 'folio-mod-roles-keycloak-capability-group' has no active members."
	ErrRebalancing      = "Consumer group 'folio-mod-roles-keycloak-capability-group' is rebalancing."
)

func GetInitialRequiredContainers() []string {
	return []string{"postgres", "kafka", "kafka-tools", "vault", "keycloak", "keycloak-internal", "kong"}
}

func GetProfiles() []string {
	return []string{"combined", "export", "search", "edge", "ecs", "ecs-single", "import"}
}

func GetDefaultProfile() string {
	return "combined"
}
