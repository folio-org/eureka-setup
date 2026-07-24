package uisvc

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-setup/eureka-cli/execsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/gitclient"
	"github.com/folio-org/eureka-setup/eureka-cli/tenantsvc"
	"github.com/go-git/go-git/v5"
)

// UIProcessor defines the composite interface for all UI-related operations
type UIProcessor interface {
	UIRepositoryCloner
	UIContainerManager
	UIPackageJSONProcessor
	UIStripesConfigProcessor
}

// UIRepositoryCloner defines the interface for cloning and updating UI repositories
type UIRepositoryCloner interface {
	CloneAndUpdateRepository(updateCloned bool) (string, error)
}

// UIContainerManager defines the interface for managing UI containers
type UIContainerManager interface {
	PrepareImage(tenantName string) (string, error)
	BuildImage(tenantName string, outputDir string) (string, error)
	DeployContainer(tenantName string, imageName string, externalPort int) error
}

// UISvc provides functionality for building and deploying UI modules
type UISvc struct {
	Action       *action.Action
	ExecSvc      execsvc.CommandRunner
	GitClient    gitclient.GitClientRunner
	DockerClient dockerclient.DockerClientRunner
	TenantSvc    tenantsvc.TenantProcessor
}

// New creates a new UISvc instance
func New(action *action.Action,
	execSvc execsvc.CommandRunner,
	gitClient gitclient.GitClientRunner,
	dockerClient dockerclient.DockerClientRunner,
	tenantSvc tenantsvc.TenantProcessor) *UISvc {
	return &UISvc{
		Action:       action,
		ExecSvc:      execSvc,
		GitClient:    gitClient,
		DockerClient: dockerClient,
		TenantSvc:    tenantSvc,
	}
}

func (us *UISvc) CloneAndUpdateRepository(updateCloned bool) (string, error) {
	slog.Info(us.Action.Name, "text", "CLONING & UPDATING PLATFORM LSP UI REPOSITORY")
	branch := us.GetStripesBranch()
	repository, err := us.GitClient.PlatformLspRepository(branch)
	if err != nil {
		return "", err
	}

	err = us.GitClient.Clone(repository)
	if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return "", err
	}

	if updateCloned {
		err = us.GitClient.ResetHardPullFromOrigin(repository)
		if err != nil {
			return "", err
		}
	}

	return repository.Dir, nil
}

func (us *UISvc) PrepareImage(tenantName string) (string, error) {
	imageName := fmt.Sprintf("platform-lsp-ui-%s", tenantName)
	if us.Action.Param.BuildImages {
		return us.buildImageFromRepository(tenantName)
	}
	if us.Action.ConfigNamespacePlatformLspUI != "" {
		if us.Action.ConfigNamespacePlatformLspUI == constant.DeprecatedUINamespace {
			slog.Warn(us.Action.Name, "text", "Configured UI namespace is deprecated, remove namespaces.platform-lsp-ui to build locally or set your own namespace populated via buildAndPushUi", "namespace", constant.DeprecatedUINamespace)
		}
		return us.DockerClient.ForcePullImage(imageName)
	}

	exists, err := us.imageExists(imageName)
	if err != nil {
		return "", err
	}
	if exists {
		slog.Info(us.Action.Name, "text", "Reusing existing local UI image, pass -b to rebuild", "image", imageName)
		return imageName, nil
	}

	return us.buildImageFromRepository(tenantName)
}

func (us *UISvc) buildImageFromRepository(tenantName string) (string, error) {
	outputDir, err := us.CloneAndUpdateRepository(us.Action.Param.UpdateCloned)
	if err != nil {
		return "", err
	}

	return us.BuildImage(tenantName, outputDir)
}

func (us *UISvc) imageExists(imageName string) (bool, error) {
	stdout, _, err := us.ExecSvc.ExecReturnOutput(exec.Command("docker", "images", "--quiet", imageName+":latest"))
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(stdout.String()) != "", nil
}

func (us *UISvc) BuildImage(tenantName string, outputDir string) (string, error) {
	// Config substitution below is destructive, work on a throwaway copy so the checkout keeps its placeholders and any user modifications
	buildDir, err := us.copySourceToTempDir(outputDir)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(buildDir); err != nil {
			slog.Warn(us.Action.Name, "text", "Removing temporary build directory was unsuccessful", "dir", buildDir, "error", err)
		}
	}()

	slog.Info(us.Action.Name, "text", "Preparing UI configs")
	err = us.PrepareStripesConfigJS(tenantName, buildDir)
	if err != nil {
		return "", err
	}

	slog.Info(us.Action.Name, "text", "Removing optional modules from stripes.modules.js")
	err = us.PrepareStripesModulesJS(buildDir)
	if err != nil {
		return "", err
	}

	err = us.PreparePackageJSON(buildDir)
	if err != nil {
		return "", err
	}

	slog.Info(us.Action.Name, "text", "Building UI image")
	finalImageName := fmt.Sprintf("platform-lsp-ui-%s", tenantName)
	err = us.ExecSvc.ExecFromDir(exec.Command("docker", "build", "--tag", finalImageName,
		"--build-arg", fmt.Sprintf("OKAPI_URL=%s", constant.KongExternalHTTP),
		"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenantName),
		"--file", "./docker/Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), buildDir)
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}

func (us *UISvc) copySourceToTempDir(sourceDir string) (string, error) {
	slog.Info(us.Action.Name, "text", "Copying UI source to a temporary build directory")
	tempDir, err := os.MkdirTemp("", "platform-lsp-ui-build-")
	if err != nil {
		return "", err
	}

	if err := os.CopyFS(tempDir, os.DirFS(sourceDir)); err != nil {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			slog.Warn(us.Action.Name, "text", "Removing temporary build directory was unsuccessful", "dir", tempDir, "error", removeErr)
		}
		return "", err
	}

	return tempDir, nil
}

func (us *UISvc) DeployContainer(tenantName string, imageName string, externalPort int) error {
	slog.Info(us.Action.Name, "text", "Deploying UI container for tenant", "tenant", tenantName)
	containerName := fmt.Sprintf("eureka-platform-lsp-ui-%s", tenantName)

	stdout, _, err := us.ExecSvc.ExecReturnOutput(exec.Command("docker", "ps", "-a",
		"--filter", fmt.Sprintf("name=^%s$", containerName),
		"--format", "{{.Names}}",
	))
	if err != nil {
		return err
	}
	if strings.TrimSpace(stdout.String()) != "" {
		slog.Info(us.Action.Name, "text", "UI container already deployed, skipping", "tenant", tenantName)
		return nil
	}

	err = us.ExecSvc.Exec(exec.Command("docker", "run", "--name", containerName,
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
	slog.Info(us.Action.Name, "text", "Connecting UI container for tenant to network", "tenant", tenantName, "network", constant.NetworkID)

	return us.ExecSvc.Exec(exec.Command("docker", "network", "connect", constant.NetworkID, containerName))
}
