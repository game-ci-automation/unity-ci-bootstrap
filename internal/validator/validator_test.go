package validator_test

import (
	"testing"

	"github.com/game-ci-automation/unity-ci-enabler/internal/validator"
)

func TestValidateUnityVersion(t *testing.T) {
	valid := []string{
		"2022.3.50f1",
		"2021.3.1f1",
		"6000.0.0f1",
	}
	for _, v := range valid {
		if !validator.ValidateUnityVersion(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{
		"",
		"invalid",
		"2022.3",
		"2022.3.50",
		"2022.3.50b1",
	}
	for _, v := range invalid {
		if validator.ValidateUnityVersion(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestValidatePlatform(t *testing.T) {
	valid := []string{
		"WebGL",
		"Android",
		"iOS",
		"StandaloneLinux64",
		"StandaloneWindows64",
	}
	for _, v := range valid {
		if !validator.ValidatePlatform(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{
		"",
		"webgl",
		"unknown",
		"PS5",
	}
	for _, v := range invalid {
		if validator.ValidatePlatform(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}
