package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/game-ci-automation/unity-ci-bootstrap/internal/cloud"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/docker"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/gcp"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/github"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/license"
	"github.com/game-ci-automation/unity-ci-bootstrap/internal/validator"
)

// GCP variant of the downloader, reduced per the serverless CI design
// (Discord-triggered): pull image → license to Secret Manager → cleanup.
// No webhook secret and no GitHub webhook registration — GitHub connects
// only to Discord; the cloud ingress is Discord Interactions.
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
	licenseContent, err := license.Read()
	if err != nil {
		log.Fatalf("failed to read license file: %v", err)
	}

	// Upload license to GCP Secret Manager.
	// Auth: Application Default Credentials via the VM's attached service
	// account (zero key files) — see internal/gcp.
	projectID := os.Getenv("PROJECT_ID")
	gcpClient, err := gcp.NewClient(projectID)
	if err != nil {
		log.Fatalf("failed to create Secret Manager client: %v", err)
	}
	var store cloud.SecretStore = gcp.NewService(gcpClient)
	if err := store.UploadLicense(licenseContent); err != nil {
		log.Fatalf("failed to upload license to Secret Manager: %v", err)
	}
	fmt.Println("License uploaded to GCP Secret Manager.")

	// Cleanup: delete license files from VM
	for _, p := range license.Paths() {
		if err := os.Remove(p); err == nil {
			fmt.Printf("Deleted license file: %s\n", p)
		}
	}

	// TODO: Kill Unity Hub process before purge — if Unity Hub is still running,
	//   apt-get purge may fail silently or skip removal.
	//   Add: exec.Command("killall", "unityhub") before purge.

	// Cleanup: uninstall Unity Hub
	uninstallCmd := exec.Command("sudo", "apt-get", "purge", "-y", "unityhub")
	uninstallCmd.Stdout = os.Stdout
	uninstallCmd.Stderr = os.Stderr
	if err := uninstallCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to uninstall Unity Hub: %v\n", err)
	} else {
		fmt.Println("Unity Hub uninstalled.")
	}

	// Print next step
	fmt.Println()
	fmt.Println("=== Bootstrap Complete ===")
	fmt.Println("Next: capture this VM as a GCP machine image, then destroy the")
	fmt.Println("ephemeral infrastructure (GCP capture script — see terraform/gcp/ephemeral).")
}
