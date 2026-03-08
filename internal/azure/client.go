package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// Client is a real Azure Key Vault client.
type Client struct {
	inner *azsecrets.Client
}

// NewClient creates a real Azure Key Vault client.
// vaultName is the Key Vault name (e.g. "my-key-vault").
// Authentication uses DefaultAzureCredential (env vars, managed identity, az CLI, etc).
func NewClient(vaultName string) (*Client, error) {
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net", vaultName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	inner, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create key vault client: %w", err)
	}

	return &Client{inner: inner}, nil
}

// SetSecret sets a secret in Azure Key Vault.
func (c *Client) SetSecret(name, value string) error {
	_, err := c.inner.SetSecret(context.Background(), name, azsecrets.SetSecretParameters{
		Value: &value,
	}, nil)
	if err != nil {
		return fmt.Errorf("set secret %q: %w", name, err)
	}
	return nil
}
