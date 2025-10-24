package vaultclient

import (
	"context"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/hashicorp/vault-client-go"
)

type VaultClient struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *VaultClient {
	return &VaultClient{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (vc *VaultClient) Create() (*vault.Client, error) {
	serverURL := vc.Action.CreateURL(constant.VaultServerPort, "")

	client, err := vault.New(vault.WithAddress(serverURL), vault.WithRequestTimeout(constant.VaultTimeout))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (vc *VaultClient) GetSecretKey(client *vault.Client, vaultRootToken string, secretPath string) (map[string]any, error) {
	err := client.SetToken(vaultRootToken)
	if err != nil {
		return nil, err
	}

	secret, err := client.Secrets.KvV2Read(context.Background(), secretPath, vault.WithMountPath("secret"))
	if err != nil {
		return nil, err
	}

	return secret.Data.Data, nil
}
