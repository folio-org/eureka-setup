package runconfig

import (
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/awssvc"
	"github.com/folio-org/eureka-cli/consortiumsvc"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/execsvc"
	"github.com/folio-org/eureka-cli/gitclient"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/interceptmodulesvc"
	"github.com/folio-org/eureka-cli/kafkasvc"
	"github.com/folio-org/eureka-cli/keycloaksvc"
	"github.com/folio-org/eureka-cli/kongsvc"
	"github.com/folio-org/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/moduleprops"
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
	*Infrastructure
	*Services
}

type Infrastructure struct {
	Action       *action.Action
	Logger       *slog.Logger
	ExecSvc      execsvc.CommandRunner
	GitClient    gitclient.GitClientRunner
	HTTPClient   httpclient.HTTPClientRunner
	DockerClient dockerclient.DockerClientRunner
	VaultClient  vaultclient.VaultClientRunner
}

type Services struct {
	AWSSvc             awssvc.AWSProcessor
	KafkaSvc           kafkasvc.KafkaProcessor
	KeycloakSvc        keycloaksvc.KeycloakProcessor
	KongSvc            kongsvc.KongProcessor
	RegistrySvc        registrysvc.RegistryProcessor
	ModuleProps        moduleprops.ModulePropsProcessor
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
	execSvc := execsvc.New(action)
	gitclient := gitclient.New(action)
	httpClient := httpclient.New(action, logger)
	dockerClient := dockerclient.New(action, execSvc)
	vaultClient := vaultclient.New(action, httpClient)
	awsSvc := awssvc.New(action)
	registrySvc := registrysvc.New(action, httpClient, awsSvc)
	moduleEnv := moduleenv.New(action)
	moduleSvc := modulesvc.New(action, httpClient, dockerClient, registrySvc, moduleEnv)
	userSvc := usersvc.New(action, httpClient)
	consortiumSvc := consortiumsvc.New(action, httpClient, userSvc)
	tenantSvc := tenantsvc.New(action, consortiumSvc)
	managementSvc := managementsvc.New(action, httpClient, tenantSvc)

	return &RunConfig{
		Infrastructure: &Infrastructure{
			Action:       action,
			Logger:       logger,
			ExecSvc:      execSvc,
			GitClient:    gitclient,
			HTTPClient:   httpClient,
			DockerClient: dockerClient,
			VaultClient:  vaultClient,
		},
		Services: &Services{
			AWSSvc:             awsSvc,
			KongSvc:            kongsvc.New(action, httpClient),
			KafkaSvc:           kafkasvc.New(action, execSvc),
			KeycloakSvc:        keycloaksvc.New(action, httpClient, vaultClient, managementSvc),
			RegistrySvc:        registrySvc,
			ModuleProps:        moduleprops.New(action),
			ModuleEnv:          moduleEnv,
			ModuleSvc:          moduleSvc,
			ManagementSvc:      managementSvc,
			TenantSvc:          tenantSvc,
			UserSvc:            userSvc,
			ConsortiumSvc:      consortiumSvc,
			UISvc:              uisvc.New(action, execSvc, gitclient, dockerClient, tenantSvc),
			SearchSvc:          searchsvc.New(action, httpClient),
			InterceptModuleSvc: interceptmodulesvc.New(action, moduleSvc, managementSvc),
		},
	}, nil
}
