package uistep

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/gitclient"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/folio-org/eureka-cli/tenantstep"
)

type UIStep struct {
	Action       *action.Action
	GitClient    *gitclient.GitClient
	DockerClient *dockerclient.DockerClient
	TenantStep   *tenantstep.TenantStep
}

func New(
	action *action.Action,
	gitClient *gitclient.GitClient,
	dockerClient *dockerclient.DockerClient,
	tenantStep *tenantstep.TenantStep,
) *UIStep {

	return &UIStep{
		Action:       action,
		GitClient:    gitClient,
		DockerClient: dockerClient,
		TenantStep:   tenantStep,
	}
}

func (us *UIStep) CloneAndUpdateUIRepository(updateCloned bool) (outputDir string, err error) {
	slog.Info(us.Action.Name, "text", "CLONING & UPDATING PLATFORM COMPLETE UI REPOSITORY")
	branch := us.GetStripesBranch()
	repository, err := us.GitClient.PlatformCompleteRepository(branch)
	if err != nil {
		return "", err
	}

	err = us.GitClient.Clone(repository)
	if err != nil {
		return "", err
	}

	if updateCloned {
		us.GitClient.ResetHardPullFromOrigin(repository)
	}

	return repository.Dir, nil
}

func (us *UIStep) PrepareUIImage(rp *runparams.RunParams, tenant string) (finalImageName string, err error) {
	imageName := fmt.Sprintf("platform-complete-ui-%s", tenant)
	if rp.BuildImages {
		outputDir, err := us.CloneAndUpdateUIRepository(rp.UpdateCloned)
		if err != nil {
			return "", err
		}

		return us.BuildImage(rp, outputDir, tenant)
	}

	finalImageName, err = us.DockerClient.ForcePullImage(imageName)
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}

func (us *UIStep) BuildImage(rp *runparams.RunParams, outputDir string, tenant string) (finalImageName string, err error) {
	finalImageName = fmt.Sprintf("platform-complete-ui-%s", tenant)

	slog.Info(us.Action.Name, "text", "Copying UI configs")
	configName := "stripes.config.js"
	helpers.CopySingleFile(us.Action, filepath.Join(outputDir, "eureka-tpl", configName), filepath.Join(outputDir, configName))

	slog.Info(us.Action.Name, "text", "PreparingUI configs")
	us.PrepareStripesConfigJS(rp, outputDir, tenant)
	us.PreparePackageJSON(outputDir, tenant)

	slog.Info(us.Action.Name, "text", "Building UI from a Dockerfile")
	err = helpers.ExecFromDir(exec.Command("docker", "build", "--tag", finalImageName,
		"--build-arg", fmt.Sprintf("OKAPI_URL=%s", constant.KongExternalHTTP),
		"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenant),
		"--file", "./docker/Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), outputDir)
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}

func (us *UIStep) DeployContainer(tenant string, imageName string, externalPort int) error {
	slog.Info(us.Action.Name, "text", fmt.Sprintf("Deploying UI container for %s tenant", tenant))
	containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", tenant)

	err := helpers.Exec(exec.Command("docker", "run", "--name", containerName,
		"--hostname", containerName,
		"--publish", fmt.Sprintf("%d:80", externalPort),
		"--restart", "unless-stopped",
		"--cpus", "1",
		"--memory", "35m",
		"--memory-swap", "-1",
		"--detach",
		imageName,
	))
	if err != nil {
		return err
	}

	slog.Info(us.Action.Name, "text", fmt.Sprintf("Connecting UI container for %s tenant to %s network", tenant, constant.NetworkID))
	err = helpers.Exec(exec.Command("docker", "network", "connect", constant.NetworkID, containerName))
	if err != nil {
		return err
	}

	return nil
}
