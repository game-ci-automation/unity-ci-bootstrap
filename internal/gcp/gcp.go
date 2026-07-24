package gcp

import "fmt"

// LicenseSecretID must match the secret created by terraform/gcp/persistent.
const LicenseSecretID = "unity-license"

type SecretManagerClient interface {
	AddSecretVersion(secretID, payload string) error
}

type Service struct {
	client SecretManagerClient
}

func NewService(client SecretManagerClient) *Service {
	return &Service{client: client}
}

// UploadLicense adds the license content as a new version of the
// unity-license secret. No webhook secret here: the Discord-triggered
// design has no cloud↔GitHub link (see serverless CI design).
func (s *Service) UploadLicense(licenseContent string) error {
	if licenseContent == "" {
		return fmt.Errorf("license content must not be empty")
	}
	return s.client.AddSecretVersion(LicenseSecretID, licenseContent)
}
