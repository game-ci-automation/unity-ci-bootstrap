package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// GHCLI abstracts the GitHub CLI for testability.
type GHCLI interface {
	Run(args ...string) (string, error)
}

// RealGHCLI calls the actual gh CLI.
type RealGHCLI struct{}

func (r *RealGHCLI) Run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	return string(out), err
}

// Client uses gh CLI to interact with GitHub.
type Client struct {
	gh GHCLI
}

// NewClient creates a Client using the real gh CLI.
func NewClient() *Client {
	return &Client{gh: &RealGHCLI{}}
}

// FetchUnityVersion reads ProjectSettings/ProjectVersion.txt from a GitHub repo
// and returns the parsed Unity version.
func (c *Client) FetchUnityVersion(owner, repo string) (string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/contents/ProjectSettings/ProjectVersion.txt", owner, repo)
	output, err := c.gh.Run("api", endpoint)
	if err != nil {
		return "", fmt.Errorf("gh api failed: %w", err)
	}

	var resp struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	return ParseUnityVersion(string(decoded))
}

// ParseUnityVersion extracts the Unity editor version from ProjectVersion.txt content.
func ParseUnityVersion(content string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "m_EditorVersion:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "m_EditorVersion:"))
			if version == "" {
				return "", fmt.Errorf("empty version in m_EditorVersion line")
			}
			return version, nil
		}
	}
	return "", fmt.Errorf("m_EditorVersion not found in ProjectVersion.txt")
}

// ParseRepoOwnerName extracts owner and repo name from a GitHub URL.
func ParseRepoOwnerName(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil || u.Host == "" {
		return "", "", fmt.Errorf("invalid GitHub URL: %q", repoURL)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL: %q", repoURL)
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	return owner, repo, nil
}
