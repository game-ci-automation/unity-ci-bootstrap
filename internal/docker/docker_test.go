package docker_test

import (
	"testing"

	"github.com/game-ci-automation/unity-ci-enabler/internal/docker"
)

func TestResolveImageTag(t *testing.T) {
	tests := []struct {
		version  string
		platform string
		want     string
		wantErr  bool
	}{
		{"2022.3.50f1", "WebGL", "unityci/editor:ubuntu-2022.3.50f1-webgl-3", false},
		{"2022.3.50f1", "Android", "unityci/editor:ubuntu-2022.3.50f1-android-3", false},
		{"2022.3.50f1", "iOS", "unityci/editor:ubuntu-2022.3.50f1-ios-3", false},
		{"2022.3.50f1", "StandaloneLinux64", "unityci/editor:ubuntu-2022.3.50f1-linux-il2cpp-3", false},
		{"2022.3.50f1", "StandaloneWindows64", "unityci/editor:ubuntu-2022.3.50f1-windows-mono-3", false},
		{"2022.3.50f1", "unknown", "", true},
	}

	for _, tt := range tests {
		got, err := docker.ResolveImageTag(tt.version, tt.platform)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ResolveImageTag(%q, %q): expected error, got nil", tt.version, tt.platform)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveImageTag(%q, %q): unexpected error: %v", tt.version, tt.platform, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ResolveImageTag(%q, %q) = %q, want %q", tt.version, tt.platform, got, tt.want)
		}
	}
}
