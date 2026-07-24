package license_test

import (
	"path/filepath"
	"testing"

	"github.com/game-ci-automation/unity-ci-bootstrap/internal/license"
)

func TestPathsForOS_Linux_Unity6(t *testing.T) {
	home := "/home/azureuser"
	paths := license.PathsForOS("linux", home)

	want := filepath.Join(home, ".config/unity3d/Unity/licenses/UnityEntitlementLicense.xml")
	for _, p := range paths {
		if p == want {
			return
		}
	}
	t.Errorf("expected Unity 6+ path %q not found in %v", want, paths)
}

func TestPathsForOS_Linux_PreUnity6(t *testing.T) {
	home := "/home/azureuser"
	paths := license.PathsForOS("linux", home)

	want := filepath.Join(home, ".local/share/unity3d/Unity/Unity_lic.ulf")
	for _, p := range paths {
		if p == want {
			return
		}
	}
	t.Errorf("expected pre-Unity 6 path %q not found in %v", want, paths)
}

func TestPathsForOS_Windows(t *testing.T) {
	paths := license.PathsForOS("windows", "")

	want := `C:\ProgramData\Unity\Unity_lic.ulf`
	for _, p := range paths {
		if p == want {
			return
		}
	}
	t.Errorf("expected path %q not found in %v", want, paths)
}

func TestPathsForOS_Darwin(t *testing.T) {
	paths := license.PathsForOS("darwin", "")

	want := "/Library/Application Support/Unity/Unity_lic.ulf"
	for _, p := range paths {
		if p == want {
			return
		}
	}
	t.Errorf("expected path %q not found in %v", want, paths)
}
