// Package license locates and reads the Unity license file on the
// bootstrap VM. Shared by all cmd/downloader-* binaries (cloud-agnostic).
package license

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Read returns the content of the first Unity license file found
// in the OS-specific well-known locations.
func Read() (string, error) {
	for _, p := range Paths() {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("license file not found; looked in: %v", Paths())
}

// Paths returns the well-known Unity license file locations for the
// current OS.
func Paths() []string {
	home, _ := os.UserHomeDir()
	return PathsForOS(runtime.GOOS, home)
}

// PathsForOS returns the well-known Unity license file locations for the
// given OS. Split out for testability.
func PathsForOS(goos, homeDir string) []string {
	switch goos {
	case "windows":
		return []string{
			`C:\ProgramData\Unity\Unity_lic.ulf`,
			`C:\ProgramData\Unity\Unity_lic.xml`,
		}
	case "darwin":
		return []string{
			"/Library/Application Support/Unity/Unity_lic.ulf",
			"/Library/Application Support/Unity/Unity_lic.xml",
		}
	default: // linux
		return []string{
			filepath.Join(homeDir, ".config/unity3d/Unity/licenses/UnityEntitlementLicense.xml"), // Unity 6+
			filepath.Join(homeDir, ".local/share/unity3d/Unity/Unity_lic.ulf"),                   // pre-Unity 6
		}
	}
}
