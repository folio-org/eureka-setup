package vaultclient

import (
	"context"

	"github.com/hashicorp/vault-client-go"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/httpclient"
)

// TODO Add testcontainers tests
// VaultClientRunner defines the interface for Vault client operations
type VaultClientRunner interface {
	Create() (*vault.Client, error)
	GetSecretKey(ctx context.Context, client *vault.Client, vaultRootToken string, secretPath string) (map[string]any, error)
}

// VaultClient provides functionality for interacting with HashiCorp Vault
type VaultClient struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
}

// New creates a new VaultClient instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner) *VaultClient {
	return &VaultClient{Action: action, HTTPClient: httpClient}
}

func (vc *VaultClient) Create() (*vault.Client, error) {
	serverURL := vc.Action.GetRequestURL(constant.VaultServerPort, "")
	client, err := vault.New(vault.WithAddress(serverURL), vault.WithRequestTimeout(constant.ContextTimeoutVaultClient))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (vc *VaultClient) GetSecretKey(ctx context.Context, client *vault.Client, vaultRootToken string, secretPath string) (map[string]any, error) {
	err := client.SetToken(vaultRootToken)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, constant.ContextTimeoutVaultClient)
	defer cancel()

	secret, err := client.Secrets.KvV2Read(ctx, secretPath, vault.WithMountPath("secret"))
	if err != nil {
		return nil, err
	}

	return secret.Data.Data, nil
}
