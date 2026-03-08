package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/game-ci-automation/unity-ci-enabler/internal/azure"
	"github.com/game-ci-automation/unity-ci-enabler/internal/docker"
	"github.com/game-ci-automation/unity-ci-enabler/internal/validator"
)

func main() {
	version := flag.String("version", "", "Unity version (e.g. 2022.3.50f1)")
	platform := flag.String("platform", "", "Target platform (e.g. WebGL, Android, iOS, StandaloneLinux64, StandaloneWindows64)")
	flag.Parse()

	if *version == "" || *platform == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Validate Unity version and platform
	if !validator.ValidateUnityVersion(*version) {
		log.Fatalf("invalid Unity version: %q", *version)
	}
	if !validator.ValidatePlatform(*platform) {
		log.Fatalf("invalid platform: %q", *platform)
	}

	// Resolve game-ci Docker image tag
	imageTag, err := docker.ResolveImageTag(*version, *platform)
	if err != nil {
		log.Fatalf("failed to resolve image tag: %v", err)
	}
	fmt.Printf("Resolved image: %s\n", imageTag)

	// Pull game-ci Docker image
	pullCmd := exec.Command("docker", "pull", imageTag)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		log.Fatalf("docker pull failed: %v", err)
	}

	// Read license file from VM
	licenseContent, err := readLicenseFile()
	if err != nil {
		log.Fatalf("failed to read license file: %v", err)
	}

	// Upload license to Azure Key Vault
	azClient, err := azure.NewClient(os.Getenv("AZURE_VAULT_NAME"))
	if err != nil {
		log.Fatalf("failed to create Azure Key Vault client: %v", err)
	}
	azSvc := azure.NewService(azClient)
	if err := azSvc.UploadLicense(licenseContent); err != nil {
		log.Fatalf("failed to upload to Azure Key Vault: %v", err)
	}
	fmt.Println("License uploaded to Azure Key Vault.")
}

func readLicenseFile() (string, error) {
	for _, p := range licensePaths() {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("license file not found; looked in: %v", licensePaths())
}

func licensePaths() []string {
	home, _ := os.UserHomeDir()
	return licensePathsForOS(runtime.GOOS, home)
}

func licensePathsForOS(goos, homeDir string) []string {
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

