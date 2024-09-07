package internal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/vault-client-go"
)

const (
	VaultServerPort int = 8200

	VaultTimeout time.Duration = 30 * time.Second
)

func GetVaultSecretKey(commandName string, enableDebug bool, vaultRootToken string, secretPath string) map[string]interface{} {
	serverUrl := fmt.Sprintf(DockerInternalUrl, VaultServerPort, "")

	client, err := vault.New(vault.WithAddress(serverUrl), vault.WithRequestTimeout(VaultTimeout))
	if err != nil {
		slog.Error(commandName, "vault.New error", "")
		panic(err)
	}

	if err := client.SetToken(vaultRootToken); err != nil {
		slog.Error(commandName, "client.SetToken error", "")
		panic(err)
	}

	secret, err := client.Secrets.KvV2Read(context.Background(), secretPath, vault.WithMountPath("secret"))
	if err != nil {
		slog.Error(commandName, "client.Secrets.KvV2Read error", "")
		panic(err)
	}

	return secret.Data.Data
}
