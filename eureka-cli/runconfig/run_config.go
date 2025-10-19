package runconfig

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/consortiumstep"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/gitclient"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/keycloakstep"
	"github.com/folio-org/eureka-cli/managementstep"
	"github.com/folio-org/eureka-cli/moduleparams"
	"github.com/folio-org/eureka-cli/modulestep"
	"github.com/folio-org/eureka-cli/registrystep"
	"github.com/folio-org/eureka-cli/searchstep"
	"github.com/folio-org/eureka-cli/tenantstep"
	"github.com/folio-org/eureka-cli/uistep"
	"github.com/folio-org/eureka-cli/userstep"
	"github.com/folio-org/eureka-cli/vaultclient"
)

// RunConfig is a central container of all dependencies
// injected through composition and dependency injection
type RunConfig struct {
	Action         *action.Action
	GitClient      *gitclient.GitClient
	HTTPClient     *httpclient.HTTPClient
	DockerClient   *dockerclient.DockerClient
	VaultClient    *vaultclient.VaultClient
	KeycloakStep   *keycloakstep.KeycloakStep
	RegistryStep   *registrystep.RegistryStep
	ModuleParams   *moduleparams.ModuleParams
	ModuleStep     *modulestep.ModuleStep
	ManagementStep *managementstep.ManagementStep
	TenantStep     *tenantstep.TenantStep
	UserStep       *userstep.UserStep
	ConsortiumStep *consortiumstep.ConsortiumStep
	UIStep         *uistep.UIStep
	SearchStep     *searchstep.SearchStep
}

func New(action *action.Action) *RunConfig {
	gitclient := gitclient.New(action)
	httpClient := httpclient.New(action)
	dockerClient := dockerclient.New(action)
	vaultClient := vaultclient.New(action, httpClient)
	registryStep := registrystep.New(action, httpClient)
	userStep := userstep.New(action, httpClient)
	consortiumStep := consortiumstep.New(action, httpClient, userStep)
	tenantstep := tenantstep.New(action, consortiumStep)

	return &RunConfig{
		Action:         action,
		GitClient:      gitclient,
		HTTPClient:     httpClient,
		DockerClient:   dockerClient,
		VaultClient:    vaultClient,
		KeycloakStep:   keycloakstep.New(action, httpClient, vaultClient),
		RegistryStep:   registryStep,
		ModuleParams:   moduleparams.New(action),
		ModuleStep:     modulestep.New(action, httpClient, dockerClient, registryStep),
		ManagementStep: managementstep.New(action, httpClient, tenantstep),
		TenantStep:     tenantstep,
		UserStep:       userStep,
		ConsortiumStep: consortiumStep,
		UIStep:         uistep.New(action, gitclient, dockerClient, tenantstep),
		SearchStep:     searchstep.New(action, httpClient),
	}
}
