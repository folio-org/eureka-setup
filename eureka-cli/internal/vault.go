package internal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/vault-client-go"
)

const (
	VaultServerPort int           = 8200
	VaultTimeout    time.Duration = 30 * time.Second

	VaultUrl string = "http://vault.eureka:8200"
)

func GetVaultSecretKey(commandName string, enableDebug bool, vaultRootToken string, secretPath string) map[string]any {
	serverUrl := fmt.Sprintf(GetGatewayHostname(), VaultServerPort, "")

	client, err := vault.New(vault.WithAddress(serverUrl), vault.WithRequestTimeout(VaultTimeout))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "vault.New error")
		panic(err)
	}

	if err := client.SetToken(vaultRootToken); err != nil {
		slog.Error(commandName, GetFuncName(), "client.SetToken error")
		panic(err)
	}

	secret, err := client.Secrets.KvV2Read(context.Background(), secretPath, vault.WithMountPath("secret"))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.Secrets.KvV2Read error")
		panic(err)
	}

	return secret.Data.Data
}
