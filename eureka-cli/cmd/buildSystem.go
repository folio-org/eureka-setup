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

// Flag tracking targeted builds
var uiOnly bool
var noCache bool

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

		if uiOnly {
            slog.Info(run.Config.Action.Name, "text", "ISOLATED FRONTEND TARGET DETECTED - Skipping infrastructure steps")

            // ⚡ AUTOMATIC NO-CACHE: Merge CLI flag with profile config option
            forceNoCache := noCache || run.Config.Action.ConfigFrontendAlwaysBuild

            if err := run.BuildCustomFrontendOnly(forceNoCache); err != nil {
                return err
            }
            slog.Info(run.Config.Action.Name, "text", "Isolated custom frontend build completed", "duration", time.Since(start))
            return nil
        }

		// Default Complete Sweep
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

func (run *Run) BuildCustomFrontendOnly(forceNoCache bool) error {
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	if run.Config.Action.ConfigFrontendPlatform == "" || run.Config.Action.ConfigFrontendURL == "" {
		return fmt.Errorf("frontend configuration missing from active profile config keys")
	}

	branchName := run.Config.Action.ConfigFrontendBranch
	if branchName == "" {
		branchName = "main"
	}

	startScript := run.Config.Action.ConfigFrontendStartScript
	if startScript == "" {
		startScript = "start"
	}

	dockerfileContent := fmt.Sprintf(`FROM node:20
WORKDIR /app
RUN git clone -b %s %s .
ENV HUSKY=0
ENV NODE_OPTIONS=--max-old-space-size=4096
RUN npm install --legacy-peer-deps
EXPOSE 3000
CMD ["npm", "run", "%s", "--", "--host", "0.0.0.0"]`, branchName, run.Config.Action.ConfigFrontendURL, startScript)

	dockerfilePath := filepath.Join(homeDir, "Dockerfile.custom-frontend")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to write dynamic frontend Dockerfile: %w", err)
	}
	defer os.Remove(dockerfilePath)

	var tenantName string
	for k := range run.Config.Action.ConfigTenants {
		tenantName = k
		break
	}
	if tenantName == "" {
		tenantName = "diku"
	}

	targetTag := fmt.Sprintf("%s/platform-lsp-ui-%s:latest", run.Config.Action.ConfigNamespacePlatformLspUI, tenantName)
	slog.Info(run.Config.Action.Name, "text", "Compiling custom frontend platform dynamically", "tag", targetTag)

	// ⚡ CACHE CONTROL
	buildArgs := []string{"build"}
	if forceNoCache {
		buildArgs = append(buildArgs, "--no-cache")
	}
	buildArgs = append(buildArgs, "-t", targetTag, "-f", "Dockerfile.custom-frontend", ".")

	buildCmd := exec.Command("docker", buildArgs...)
	if err := run.Config.ExecSvc.ExecFromDir(buildCmd, homeDir); err != nil {
		return fmt.Errorf("failed compiling custom workspace image container: %w", err)
	}

	return nil
}

func (run *Run) BuildSystem() error {
	if run.Config.Action.ConfigFrontendPlatform != "" && run.Config.Action.ConfigFrontendURL != "" {
		slog.Info(run.Config.Action.Name, "text", "DYNAMIC FRONTEND PLATFORM DETECTED IN FULL SWEEP")

		// ⚡ AUTOMATIC NO-CACHE: Merge CLI flag with profile config option here too
		forceNoCache := noCache || run.Config.Action.ConfigFrontendAlwaysBuild

		if err := run.BuildCustomFrontendOnly(forceNoCache); err != nil {
			return err
		}
	}

	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "BUILDING SYSTEM IMAGES")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	return run.Config.ExecSvc.ExecFromDir(exec.Command("docker", subCommand...), homeDir)
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
	buildSystemCmd.Flags().BoolVar(&uiOnly, "ui-only", false, "Compile only the dynamic custom frontend platform image")
	buildSystemCmd.Flags().BoolVar(&noCache, action.NoCache.Long, false, action.NoCache.Description)
}