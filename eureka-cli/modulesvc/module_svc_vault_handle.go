package modulesvc

import (
	"context"
	"encoding/binary"
	"log/slog"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/moby/moby/client"
)

// ModuleVaultHandler defines the interface for module Vault operations
type ModuleVaultHandler interface {
	GetVaultRootToken(client *client.Client) (string, error)
}

func (ms *ModuleSvc) GetVaultRootToken(dockerClient *client.Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutVaultContainerLogs)
	defer cancel()

	logStream, err := dockerClient.ContainerLogs(ctx, constant.VaultContainer, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", err
	}
	defer helpers.CloseReader(logStream)

	buffer := make([]byte, constant.DockerLogHeaderSize)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			return "", err
		}

		count := binary.BigEndian.Uint32(buffer[constant.DockerLogSizeOffset:])
		rawLogLine := make([]byte, count)
		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
		}

		parsedLogLine := string(rawLogLine)
		if strings.Contains(parsedLogLine, constant.VaultRootTokenPattern) {
			vaultRootToken := helpers.GetVaultRootTokenFromLogs(parsedLogLine)

			return vaultRootToken, nil
		}
	}
}
