package modulesvc

import (
	"context"
	"encoding/binary"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
)

type ModuleVaultHandler interface {
	GetVaultRootToken(client *client.Client) (string, error)
}

func (ms *ModuleSvc) GetVaultRootToken(client *client.Client) (string, error) {
	logStream, err := client.ContainerLogs(context.Background(), constant.VaultContainer, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", err
	}
	defer helpers.CloseReader(logStream)

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			return "", err
		}

		count := binary.BigEndian.Uint32(buffer[4:])
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
