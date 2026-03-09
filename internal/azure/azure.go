package azure

import "fmt"

type KeyVaultClient interface {
	SetSecret(name, value string) error
}

type Service struct {
	client KeyVaultClient
}

func NewService(client KeyVaultClient) *Service {
	return &Service{client: client}
}

func (s *Service) UploadLicense(licenseContent string) error {
	if licenseContent == "" {
		return fmt.Errorf("license content must not be empty")
	}
	return s.client.SetSecret("UNITY-LICENSE", licenseContent)
}

func (s *Service) UploadWebhookSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("webhook secret must not be empty")
	}
	return s.client.SetSecret("WEBHOOK-SECRET", secret)
}
