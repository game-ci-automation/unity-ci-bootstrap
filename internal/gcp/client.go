package gcp

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Client is a real GCP Secret Manager client.
type Client struct {
	inner     *secretmanager.Client
	projectID string
}

// NewClient creates a real GCP Secret Manager client.
// Authentication uses Application Default Credentials — on the bootstrap VM
// this resolves to the attached service account via the metadata server
// (zero key files; GCP's equivalent of Azure Managed Identity).
func NewClient(projectID string) (*Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID must not be empty")
	}

	inner, err := secretmanager.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create secret manager client: %w", err)
	}

	return &Client{inner: inner, projectID: projectID}, nil
}

// AddSecretVersion adds a new version to an existing secret.
// The secret container itself is created by terraform/gcp/persistent;
// the VM's service account only needs roles/secretmanager.secretVersionAdder
// on that secret (least privilege — it cannot even read versions back).
func (c *Client) AddSecretVersion(secretID, payload string) error {
	_, err := c.inner.AddSecretVersion(context.Background(), &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", c.projectID, secretID),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(payload),
		},
	})
	if err != nil {
		return fmt.Errorf("add version to secret %q: %w", secretID, err)
	}
	return nil
}
