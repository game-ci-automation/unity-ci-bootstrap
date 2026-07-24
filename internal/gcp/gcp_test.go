package gcp_test

import (
	"errors"
	"testing"

	"github.com/game-ci-automation/unity-ci-bootstrap/internal/gcp"
)

type mockSecretManagerClient struct {
	uploadedID      string
	uploadedPayload string
	shouldFail      bool
}

func (m *mockSecretManagerClient) AddSecretVersion(secretID, payload string) error {
	if m.shouldFail {
		return errors.New("add secret version failed")
	}
	m.uploadedID = secretID
	m.uploadedPayload = payload
	return nil
}

func TestUploadLicense(t *testing.T) {
	mock := &mockSecretManagerClient{}
	svc := gcp.NewService(mock)

	err := svc.UploadLicense("license-content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.uploadedID != "unity-license" {
		t.Errorf("secret id = %q, want %q", mock.uploadedID, "unity-license")
	}
	if mock.uploadedPayload != "license-content" {
		t.Errorf("secret payload = %q, want %q", mock.uploadedPayload, "license-content")
	}
}

func TestUploadLicense_ClientError(t *testing.T) {
	mock := &mockSecretManagerClient{shouldFail: true}
	svc := gcp.NewService(mock)

	err := svc.UploadLicense("license-content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadLicense_EmptyLicense(t *testing.T) {
	mock := &mockSecretManagerClient{}
	svc := gcp.NewService(mock)

	err := svc.UploadLicense("")
	if err == nil {
		t.Fatal("expected error for empty license, got nil")
	}
}
