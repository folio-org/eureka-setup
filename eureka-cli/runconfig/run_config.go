package runconfig

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/consortiumsvc"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/gitclient"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/kafkasvc"
	"github.com/folio-org/eureka-cli/keycloaksvc"
	"github.com/folio-org/eureka-cli/managementsvc"
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
	Action        *action.Action
	GitClient     *gitclient.GitClient
	HTTPClient    *httpclient.HTTPClient
	DockerClient  *dockerclient.DockerClient
	VaultClient   *vaultclient.VaultClient
	KafkaSvc      *kafkasvc.KafkaSvc
	KeycloakSvc   *keycloaksvc.KeycloakSvc
	RegistrySvc   *registrysvc.RegistrySvc
	ModuleParams  *moduleparams.ModuleParams
	ModuleSvc     *modulesvc.ModuleSvc
	ManagementSvc *managementsvc.ManagementSvc
	TenantSvc     *tenantsvc.TenantSvc
	UserSvc       *usersvc.UserSvc
	ConsortiumSvc *consortiumsvc.ConsortiumSvc
	UISvc         *uisvc.UISvc
	SearchSvc     *searchsvc.SearchSvc
}

func New(action *action.Action) *RunConfig {
	gitclient := gitclient.New(action)
	httpClient := httpclient.New(action)
	dockerClient := dockerclient.New(action)
	vaultClient := vaultclient.New(action, httpClient)
	kafkaSvc := kafkasvc.New(action)
	registrySvc := registrysvc.New(action, httpClient)
	userSvc := usersvc.New(action, httpClient)
	consortiumSvc := consortiumsvc.New(action, httpClient, userSvc)
	tenantSvc := tenantsvc.New(action, consortiumSvc)
	managementSvc := managementsvc.New(action, httpClient, tenantSvc)

	return &RunConfig{
		Action:        action,
		GitClient:     gitclient,
		HTTPClient:    httpClient,
		DockerClient:  dockerClient,
		VaultClient:   vaultClient,
		KafkaSvc:      kafkaSvc,
		KeycloakSvc:   keycloaksvc.New(action, httpClient, vaultClient, managementSvc),
		RegistrySvc:   registrySvc,
		ModuleParams:  moduleparams.New(action),
		ModuleSvc:     modulesvc.New(action, httpClient, dockerClient, registrySvc),
		ManagementSvc: managementSvc,
		TenantSvc:     tenantSvc,
		UserSvc:       userSvc,
		ConsortiumSvc: consortiumSvc,
		UISvc:         uisvc.New(action, gitclient, dockerClient, tenantSvc),
		SearchSvc:     searchsvc.New(action, httpClient),
	}
}
