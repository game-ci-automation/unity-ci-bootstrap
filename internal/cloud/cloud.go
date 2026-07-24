// Package cloud defines the cross-cloud contracts implemented by
// internal/azure and internal/gcp. Each cmd/downloader-* binary imports
// only its own cloud implementation, so the other cloud's SDK is never
// compiled in.
package cloud

// SecretStore uploads the Unity license to a cloud secret store
// (Azure Key Vault / GCP Secret Manager).
type SecretStore interface {
	UploadLicense(licenseContent string) error
}
