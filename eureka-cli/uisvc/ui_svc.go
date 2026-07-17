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
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
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
	if us.Action.ConfigFrontendPlatform != "" {
        targetTag := helpers.ResolveUIPlatformTag(us.Action.ConfigNamespacePlatformLspUI, tenantName)

		// Trigger compilation if global build (-b), targeted build UI flag, OR always-build config is active
		if us.Action.Param.BuildImages || us.Action.Param.BuildUI || us.Action.ConfigFrontendAlwaysBuild {
			slog.Info(us.Action.Name, "text", "Compiling custom frontend platform via native deploy hook", "tag", targetTag)

			// ⚡ AUTOMATIC NO-CACHE: Force no-cache if flag is passed OR always-build is true in configuration
			forceNoCache := us.Action.Param.NoCache || us.Action.ConfigFrontendAlwaysBuild

			if err := us.CompileCustomImage(tenantName, forceNoCache); err != nil {
				return "", err
			}
			return targetTag, nil
		}

		slog.Info(us.Action.Name, "text", "Using cached custom platform image target", "tag", targetTag)
		return targetTag, nil
	}

	// Legacy Native Fallback
	imageName := fmt.Sprintf("platform-lsp-ui-%s", tenantName)
	if us.Action.Param.BuildImages {
		outputDir, err := us.CloneAndUpdateRepository(us.Action.Param.UpdateCloned)
		if err != nil {
			return "", err
		}

		return us.BuildImage(tenantName, outputDir)
	}

	finalImageName, err := us.DockerClient.ForcePullImage(imageName)
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}

func (us *UISvc) BuildImage(tenantName string, outputDir string) (string, error) {
	slog.Info(us.Action.Name, "text", "Preparing UI configs")
	err := us.PrepareStripesConfigJS(tenantName, outputDir)
	if err != nil {
		return "", err
	}

	slog.Info(us.Action.Name, "text", "Removing optional modules from stripes.modules.js")
	err = us.PrepareStripesModulesJS(outputDir)
	if err != nil {
		return "", err
	}

	err = us.PreparePackageJSON(outputDir)
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
	), outputDir)
	if err != nil {
		return "", err
	}

	return finalImageName, nil
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

	internalPort := 80
	memoryLimit := "35m"

	// If this is a custom node platform development box, expand boundaries
	if us.Action.ConfigFrontendPlatform != "" {
		internalPort = 3000
		memoryLimit = "1024m" // Give the Node container 1GB of headroom to build/watch assets
	}

	err = us.ExecSvc.Exec(exec.Command("docker", "run", "--name", containerName,
		"--hostname", containerName,
		"--publish", fmt.Sprintf("%d:%d", externalPort, internalPort),
		"--restart", "unless-stopped",
		"--cpus", "1",
		"--memory", memoryLimit,
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

func (us *UISvc) CompileCustomImage(tenantName string, noCache bool) error {
	repoURL := us.Action.ConfigFrontendURL
	platformName := us.Action.ConfigFrontendPlatform
	namespace := us.Action.ConfigNamespacePlatformLspUI

	if repoURL == "" || platformName == "" {
		return fmt.Errorf("frontend configuration missing from active profile config keys")
	}

	branchName := us.Action.ConfigFrontendBranch
	if branchName == "" {
		branchName = "main"
	}

	startScript := us.Action.ConfigFrontendStartScript
	if startScript == "" {
		startScript = "start"
	}

    localImageTag := helpers.ResolveUIPlatformTag(namespace, tenantName)

	dockerfileContent, err := helpers.GenerateFrontendDockerfile(branchName, repoURL, startScript)
	if err != nil {
		return fmt.Errorf("failed compiling custom frontend payload configurations: %w", err)
	}

	dockerfilePath := "Dockerfile.custom-frontend"
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed creating transient dockerfile spec: %w", err)
	}
	defer os.Remove(dockerfilePath)

	buildArgs := []string{"build"}
	if noCache {
		buildArgs = append(buildArgs, "--no-cache")
	}
	buildArgs = append(buildArgs, "-t", localImageTag, "-f", dockerfilePath, ".")

	cmd := exec.Command("docker", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return us.ExecSvc.Exec(cmd)
}