package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformAzureVM(t *testing.T) {
	t.Parallel()

	opts := &terraform.Options{
		TerraformDir:    "../terraform/ephemeral",
		TerraformBinary: `C:\Users\blitz\AppData\Local\Microsoft\WinGet\Packages\Hashicorp.Terraform_Microsoft.Winget.Source_8wekyb3d8bbwe\terraform.exe`,
		Vars: map[string]interface{}{
			"admin_password": "TestP@ssw0rd123!",
		},
	}

	// Destroy after test completes (pass or fail)
	defer terraform.Destroy(t, opts)

	// Apply Terraform
	terraform.InitAndApply(t, opts)

	// Verify outputs exist
	publicIP := terraform.Output(t, opts, "public_ip_address")
	require.NotEmpty(t, publicIP, "public_ip_address output should not be empty")

	vmID := terraform.Output(t, opts, "vm_id")
	require.NotEmpty(t, vmID, "vm_id output should not be empty")

	novncURL := terraform.Output(t, opts, "novnc_url")
	require.NotEmpty(t, novncURL, "novnc_url output should not be empty")

	t.Logf("VM public IP: %s", publicIP)
	t.Logf("noVNC URL: %s", novncURL)

	// Wait for cloud-init to complete (up to 10 minutes)
	t.Log("Waiting for cloud-init to complete...")
	novncReachable := waitForHTTP(novncURL, 10*time.Minute)
	assert.True(t, novncReachable, "noVNC should be accessible on port 6080 after cloud-init")
}

// waitForHTTP polls the given URL until it responds with 200 or timeout.
func waitForHTTP(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/vnc.html", url))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(30 * time.Second)
	}
	return false
}
