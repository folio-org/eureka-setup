package dockerclient

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/execsvc"
	"github.com/j011195/eureka-setup/eureka-cli/field"
)

// TODO Add testcontainers tests
// DockerClientRunner defines the interface for Docker client operations
type DockerClientRunner interface {
	Create() (*client.Client, error)
	Close(client *client.Client)
	PushImage(namespace string, imageName string) error
	ForcePullImage(imageName string) (finalImageName string, err error)
}

// DockerClient provides functionality for Docker operations
type DockerClient struct {
	Action  *action.Action
	ExecSvc execsvc.CommandRunner
}

// New creates a new DockerClient instance
func New(action *action.Action, execSvc execsvc.CommandRunner) *DockerClient {
	return &DockerClient{Action: action, ExecSvc: execSvc}
}

func (dc *DockerClient) Create() (*client.Client, error) {
	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerAPIVersion)
	defer cancel()

	newClient.NegotiateAPIVersion(ctx)

	return newClient, nil
}

func (dc *DockerClient) Close(client *client.Client) {
	_ = client.Close()
}

func (dc *DockerClient) PushImage(namespace string, imageName string) error {
	slog.Info(dc.Action.Name, "text", "PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB")
	finalImageName := fmt.Sprintf("%s/%s", namespace, imageName)

	slog.Info(dc.Action.Name, "text", "Tagging platform complete UI image")
	err := dc.ExecSvc.Exec(exec.Command("docker", "tag", imageName, finalImageName))
	if err != nil {
		return err
	}

	slog.Info(dc.Action.Name, "text", "Pushing new platform complete UI image to Docker Hub")
	err = dc.ExecSvc.Exec(exec.Command("docker", "push", finalImageName))
	if err != nil {
		return err
	}

	return nil
}

func (dc *DockerClient) ForcePullImage(imageName string) (finalImageName string, err error) {
	slog.Info(dc.Action.Name, "text", "PULLING PLATFORM COMPLETE UI IMAGE FROM DOCKER HUB")
	if !action.IsSet(field.NamespacesPlatformCompleteUI) {
		return "", errors.ImageKeyNotSet(imageName, field.NamespacesPlatformCompleteUI)
	}

	finalImageName = fmt.Sprintf("%s/%s", dc.Action.ConfigNamespacePlatformCompleteUI, imageName)
	slog.Info(dc.Action.Name, "text", "Removing old platform complete UI image")
	err = dc.ExecSvc.Exec(exec.Command("docker", "image", "rm", "--force", finalImageName))
	if err != nil {
		return "", err
	}

	slog.Info(dc.Action.Name, "text", "Pulling new platform complete UI image from Docker Hub")
	err = dc.ExecSvc.Exec(exec.Command("docker", "image", "pull", finalImageName))
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}
