package runconfig

import (
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/awssvc"
	"github.com/folio-org/eureka-cli/consortiumsvc"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/gitclient"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/interceptmodulesvc"
	"github.com/folio-org/eureka-cli/kafkasvc"
	"github.com/folio-org/eureka-cli/keycloaksvc"
	"github.com/folio-org/eureka-cli/kongsvc"
	"github.com/folio-org/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/moduleparams"
	"github.com/folio-org/eureka-cli/modulesvc"
	"github.com/folio-org/eureka-cli/registrysvc"
	"github.com/folio-org/eureka-cli/searchsvc"
	"github.com/folio-org/eureka-cli/tenantsvc"
	"github.com/folio-org/eureka-cli/uisvc"
	"github.com/folio-org/eureka-cli/usersvc"
	"github.com/folio-org/eureka-cli/vaultclient"
)

// RunConfig is a central container of all dependencies (services)
// manually injected through composition and dependency injection
type RunConfig struct {
	// Core Infrastructure
	Action             *action.Action
	Logger             *slog.Logger
	GitClient          gitclient.GitClientRunner
	HTTPClient         httpclient.HTTPClientRunner
	DockerClient       dockerclient.DockerClientRunner
	VaultClient        vaultclient.VaultClientRunner
	AWSSvc             awssvc.AWSProcessor
	KafkaSvc           kafkasvc.KafkaProcessor
	KeycloakSvc        keycloaksvc.KeycloakProcessor
	KongSvc            kongsvc.KongProcessor
	RegistrySvc        registrysvc.RegistryProcessor
	ModuleParams       moduleparams.ModuleParamsProcessor
	ModuleEnv          moduleenv.ModuleEnvProcessor
	ModuleSvc          modulesvc.ModuleProcessor
	ManagementSvc      managementsvc.ManagementProcessor
	TenantSvc          tenantsvc.TenantProcessor
	UserSvc            usersvc.UserProcessor
	ConsortiumSvc      consortiumsvc.ConsortiumProcessor
	UISvc              uisvc.UIProcessor
	SearchSvc          searchsvc.SearchProcessor
	InterceptModuleSvc interceptmodulesvc.InterceptModuleProcessor
}

func New(action *action.Action, logger *slog.Logger) (*RunConfig, error) {
	if action == nil {
		return nil, errors.ActionNil()
	}
	if logger == nil {
		return nil, errors.LoggerNil()
	}

	// Create infrastructure
	gitclient := gitclient.New(action)
	httpClient := httpclient.New(action, logger)
	dockerClient := dockerclient.New(action)
	vaultClient := vaultclient.New(action, httpClient)

	// Create services
	awsSvc := awssvc.New(action)
	registrySvc := registrysvc.New(action, httpClient, awsSvc)
	moduleEnv := moduleenv.New(action)
	moduleSvc := modulesvc.New(action, httpClient, dockerClient, registrySvc, moduleEnv)
	userSvc := usersvc.New(action, httpClient)
	consortiumSvc := consortiumsvc.New(action, httpClient, userSvc)
	tenantSvc := tenantsvc.New(action, consortiumSvc)
	managementSvc := managementsvc.New(action, httpClient, tenantSvc)

	return &RunConfig{
		Action:             action,
		Logger:             logger,
		GitClient:          gitclient,
		HTTPClient:         httpClient,
		DockerClient:       dockerClient,
		VaultClient:        vaultClient,
		AWSSvc:             awsSvc,
		KongSvc:            kongsvc.New(action, httpClient),
		KafkaSvc:           kafkasvc.New(action),
		KeycloakSvc:        keycloaksvc.New(action, httpClient, vaultClient, managementSvc),
		RegistrySvc:        registrySvc,
		ModuleParams:       moduleparams.New(action),
		ModuleEnv:          moduleEnv,
		ModuleSvc:          moduleSvc,
		ManagementSvc:      managementSvc,
		TenantSvc:          tenantSvc,
		UserSvc:            userSvc,
		ConsortiumSvc:      consortiumSvc,
		UISvc:              uisvc.New(action, gitclient, dockerClient, tenantSvc),
		SearchSvc:          searchsvc.New(action, httpClient),
		InterceptModuleSvc: interceptmodulesvc.New(action, moduleSvc, managementSvc),
	}, nil
}
