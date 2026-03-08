package azure_test

import (
	"errors"
	"testing"

	"github.com/game-ci-automation/unity-ci-enabler/internal/azure"
)

type mockKeyVaultClient struct {
	uploadedName  string
	uploadedValue string
	shouldFail    bool
}

func (m *mockKeyVaultClient) SetSecret(name, value string) error {
	if m.shouldFail {
		return errors.New("set secret failed")
	}
	m.uploadedName = name
	m.uploadedValue = value
	return nil
}

func TestUploadLicense(t *testing.T) {
	mock := &mockKeyVaultClient{}
	svc := azure.NewService(mock)

	err := svc.UploadLicense("license-content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.uploadedName != "UNITY-LICENSE" {
		t.Errorf("secret name = %q, want %q", mock.uploadedName, "UNITY-LICENSE")
	}
	if mock.uploadedValue != "license-content" {
		t.Errorf("secret value = %q, want %q", mock.uploadedValue, "license-content")
	}
}

func TestUploadLicense_ClientError(t *testing.T) {
	mock := &mockKeyVaultClient{shouldFail: true}
	svc := azure.NewService(mock)

	err := svc.UploadLicense("license-content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadLicense_EmptyLicense(t *testing.T) {
	mock := &mockKeyVaultClient{}
	svc := azure.NewService(mock)

	err := svc.UploadLicense("")
	if err == nil {
		t.Fatal("expected error for empty license, got nil")
	}
}
