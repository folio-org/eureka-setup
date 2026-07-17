package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/gitrepository"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// buildSystemCmd represents the buildSystem command
var buildSystemCmd = &cobra.Command{
	Use:   "buildSystem",
	Short: "Build system",
	Long:  `Build system images.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		run, err := New(action.BuildSystem)
		if err != nil {
			return err
		}
		if err := run.CloneUpdateRepositories(); err != nil {
			return err
		}
		if err := run.BuildSystem(); err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Command completed", "duration", time.Since(start))

		return nil
	},
}

func (run *Run) CloneUpdateRepositories() error {
	slog.Info(run.Config.Action.Name, "text", "CLONING & UPDATING REPOSITORIES")
	kongRepository, err := run.Config.GitClient.KongRepository()
	if err != nil {
		return err
	}

	keycloakRepository, err := run.Config.GitClient.KeycloakRepository()
	if err != nil {
		return err
	}

	repositories := []*gitrepository.GitRepository{kongRepository, keycloakRepository}

	slog.Info(run.Config.Action.Name, "text", "Cloning repositories", "repositories", repositories)
	for _, repository := range repositories {
		if err := run.Config.GitClient.Clone(repository); err != nil {
			if errors.Is(err, git.ErrRepositoryAlreadyExists) {
				slog.Info(run.Config.Action.Name, "text", "Repository already cloned, skipping", "label", repository.Label)
			} else {
				slog.Warn(run.Config.Action.Name, "text", "Cloning was unsuccessful", "error", err)
			}
		}
	}

	if params.UpdateCloned {
		slog.Info(run.Config.Action.Name, "text", "Updating repositories", "repositories", repositories)
		for _, repository := range repositories {
			if err := run.Config.GitClient.ResetHardPullFromOrigin(repository); err != nil {
				return err
			}
		}
	}

	return nil
}

func (run *Run) BuildSystem() error {
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	// --- DYNAMIC FRONTEND PLATFORM CONTAINER GENERATOR ---
	if run.Config.Action.ConfigFrontendPlatform != "" && run.Config.Action.ConfigFrontendURL != "" {
		slog.Info(run.Config.Action.Name, "text", "DYNAMIC FRONTEND PLATFORM SETUP", "platform", run.Config.Action.ConfigFrontendPlatform)

		branchName := run.Config.Action.ConfigFrontendBranch
		if branchName == "" {
			branchName = "main"
		}

		startScript := run.Config.Action.ConfigFrontendStartScript
		if startScript == "" {
			startScript = "start"
		}

        // Generate the specification dynamically
		dockerfileContent := fmt.Sprintf(`FROM node:20
WORKDIR /app
RUN git clone -b %s %s .
ENV HUSKY=0
RUN npm install --legacy-peer-deps
EXPOSE 3000
CMD ["npm", "run", "%s"]`, branchName, run.Config.Action.ConfigFrontendURL, startScript)

		dockerfilePath := filepath.Join(homeDir, "Dockerfile.custom-frontend")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
			return fmt.Errorf("failed to write dynamic frontend Dockerfile: %w", err)
		}

		// Resolve targeted application workspace tags (e.g. bkadirkhodjaev/platform-lsp-ui-ill:latest)
		var tenantName string
		for k := range run.Config.Action.ConfigTenants {
			tenantName = k // Resolves target mapping key (e.g., "ill")
			break
		}
		if tenantName == "" {
			tenantName = "diku" // Graceful default fallback
		}

		targetTag := fmt.Sprintf("%s/platform-lsp-ui-%s:latest", run.Config.Action.ConfigNamespacePlatformLspUI, tenantName)
		slog.Info(run.Config.Action.Name, "text", "Compiling custom frontend platform via container engine layer", "tag", targetTag)

		buildCmd := exec.Command("docker", "build", "-t", targetTag, "-f", "Dockerfile.custom-frontend", ".")
		if err := run.Config.ExecSvc.ExecFromDir(buildCmd, homeDir); err != nil {
			return fmt.Errorf("failed compiling custom workspace image container: %w", err)
		}
	}
	// --- END OF PATCH ---

	slog.Info(run.Config.Action.Name, "text", "BUILDING SYSTEM IMAGES")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	return run.Config.ExecSvc.ExecFromDir(exec.Command("docker", subCommand...), homeDir)
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
}