package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/game-ci-automation/unity-ci-bootstrap/internal/azure"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/docker"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/github"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/validator"
)

func main() {
	version := flag.String("version", "", "Unity version (e.g. 2022.3.50f1) — auto-detected from repo if omitted")
	platform := flag.String("platform", os.Getenv("PLATFORM"), "Target platform (e.g. WebGL, Android, iOS, StandaloneLinux64, StandaloneWindows64)")
	repoURL := flag.String("repo", os.Getenv("REPO_URL"), "GitHub repository URL (defaults to REPO_URL env var)")
	flag.Parse()

	if *platform == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Auto-detect Unity version from GitHub repo if not provided
	if *version == "" {
		if *repoURL == "" {
			log.Fatal("--version or --repo (or REPO_URL env var) is required")
		}
		owner, repo, err := github.ParseRepoOwnerName(*repoURL)
		if err != nil {
			log.Fatalf("invalid repo URL: %v", err)
		}
		ghClient := github.NewClient()
		detected, err := ghClient.FetchUnityVersion(owner, repo)
		if err != nil {
			log.Fatalf("failed to auto-detect Unity version: %v", err)
		}
		*version = detected
		fmt.Printf("Auto-detected Unity version: %s\n", *version)
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
	vaultName := os.Getenv("KEY_VAULT_NAME")
	azClient, err := azure.NewClient(vaultName)
	if err != nil {
		log.Fatalf("failed to create Azure Key Vault client: %v", err)
	}
	azSvc := azure.NewService(azClient)
	if err := azSvc.UploadLicense(licenseContent); err != nil {
		log.Fatalf("failed to upload license to Key Vault: %v", err)
	}
	fmt.Println("License uploaded to Azure Key Vault.")

	// Generate and upload webhook secret
	webhookSecret, err := generateSecret(32)
	if err != nil {
		log.Fatalf("failed to generate webhook secret: %v", err)
	}
	if err := azSvc.UploadWebhookSecret(webhookSecret); err != nil {
		log.Fatalf("failed to upload webhook secret to Key Vault: %v", err)
	}
	fmt.Println("Webhook secret uploaded to Azure Key Vault.")

	// Register GitHub webhook
	functionURL := os.Getenv("FUNCTION_URL")
	if functionURL == "" {
		log.Fatal("FUNCTION_URL env var is required for webhook registration")
	}
	owner, repo, err := github.ParseRepoOwnerName(*repoURL)
	if err != nil {
		log.Fatalf("invalid repo URL: %v", err)
	}
	ghClient := github.NewClient()
	if err := ghClient.RegisterWebhook(owner, repo, functionURL, webhookSecret); err != nil {
		log.Fatalf("failed to register webhook: %v", err)
	}
	fmt.Println("GitHub webhook registered.")

	// Cleanup: delete license files from VM
	for _, p := range licensePaths() {
		if err := os.Remove(p); err == nil {
			fmt.Printf("Deleted license file: %s\n", p)
		}
	}

	// Cleanup: uninstall Unity Hub
	uninstallCmd := exec.Command("sudo", "apt-get", "purge", "-y", "unityhub")
	uninstallCmd.Stdout = os.Stdout
	uninstallCmd.Stderr = os.Stderr
	if err := uninstallCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to uninstall Unity Hub: %v\n", err)
	} else {
		fmt.Println("Unity Hub uninstalled.")
	}

	// Print image capture instructions
	rg := os.Getenv("RESOURCE_GROUP_NAME")
	gallery := os.Getenv("IMAGE_GALLERY_NAME")
	imageDef := os.Getenv("IMAGE_DEFINITION_NAME")

	fmt.Println()
	fmt.Println("=== Bootstrap Complete ===")
	fmt.Println("Run these commands from your local machine:")
	fmt.Println()
	fmt.Printf("  az vm deallocate --resource-group %s --name unity-ci-vm\n", rg)
	fmt.Println()
	fmt.Printf("  az sig image-version create \\\n")
	fmt.Printf("    --resource-group %s \\\n", rg)
	fmt.Printf("    --gallery-name %s \\\n", gallery)
	fmt.Printf("    --gallery-image-definition %s \\\n", imageDef)
	fmt.Printf("    --gallery-image-version 1.0.0 \\\n")
	fmt.Printf("    --virtual-machine $(az vm show -g %s -n unity-ci-vm --query id -o tsv)\n", rg)
	fmt.Println()
	fmt.Println("Then destroy the ephemeral VM:")
	fmt.Println("  cd terraform/ephemeral && terraform destroy")
}

func generateSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
