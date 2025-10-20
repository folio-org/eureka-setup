package dockerclient

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/viper"
)

type DockerClient struct {
	Action *action.Action
}

func New(action *action.Action) *DockerClient {
	return &DockerClient{
		Action: action,
	}
}

func (dc *DockerClient) Create() (*client.Client, error) {
	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	newClient.NegotiateAPIVersion(context.Background())

	return newClient, nil
}

func (dc *DockerClient) Close(client *client.Client) {
	_ = client.Close()
}

func (dc *DockerClient) PushImage(namespace string, imageName string) error {
	slog.Info(dc.Action.Name, "text", "PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB")
	finalImageName := fmt.Sprintf("%s/%s", namespace, imageName)

	slog.Info(dc.Action.Name, "text", "Tagging platform complete UI image")
	err := helpers.Exec(exec.Command("docker", "tag", imageName, finalImageName))
	if err != nil {
		return err
	}

	slog.Info(dc.Action.Name, "text", "Pushing new platform complete UI image to Docker Hub")
	err = helpers.Exec(exec.Command("docker", "push", finalImageName))
	if err != nil {
		return err
	}

	return nil
}

func (dc *DockerClient) ForcePullImage(imageName string) (finalImageName string, err error) {
	slog.Info(dc.Action.Name, "text", "PULLING PLATFORM COMPLETE UI IMAGE FROM DOCKER HUB")
	if !viper.IsSet(field.NamespacesPlatformCompleteUI) {
		return "", fmt.Errorf("cannot run %s image key %s is not set in current config file", imageName, field.NamespacesPlatformCompleteUI)
	}

	finalImageName = fmt.Sprintf("%s/%s", viper.GetString(field.NamespacesPlatformCompleteUI), imageName)

	slog.Info(dc.Action.Name, "text", "Removing old platform complete UI image")
	err = helpers.Exec(exec.Command("docker", "image", "rm", "--force", finalImageName))
	if err != nil {
		return "", err
	}

	slog.Info(dc.Action.Name, "text", "Pulling new platform complete UI image from Docker Hub")
	err = helpers.Exec(exec.Command("docker", "image", "pull", finalImageName))
	if err != nil {
		return "", err
	}

	return finalImageName, nil
}
