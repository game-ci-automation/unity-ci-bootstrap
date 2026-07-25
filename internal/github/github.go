package github

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// GHCLI abstracts the GitHub CLI for testability.
// Only the webhook registration flow (Azure legacy) needs it.
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

// FileFetcher abstracts a plain HTTP file fetch for testability.
// Authenticated reports whether requests carry a token — callers pick the
// endpoint accordingly (see FetchUnityVersion).
type FileFetcher interface {
	Fetch(url string) (string, error)
	Authenticated() bool
}

// HTTPFetcher fetches with net/http. Token is optional: empty works for
// public repos; set it (GitHub PAT) to reach private repos.
type HTTPFetcher struct {
	Token string
}

func (h *HTTPFetcher) Authenticated() bool {
	return h.Token != ""
}

func (h *HTTPFetcher) Fetch(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	if h.Token != "" {
		req.Header.Set("Authorization", "token "+h.Token)
		// Makes the GitHub contents API return the file body directly.
		req.Header.Set("Accept", "application/vnd.github.raw")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", rawURL, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Client interacts with GitHub.
type Client struct {
	gh      GHCLI
	fetcher FileFetcher
}

// NewClient creates a Client with the real gh CLI and HTTP fetcher.
// GITHUB_TOKEN (or GH_TOKEN, the gh CLI convention used on the Azure VM)
// is picked up when present — optional, only needed for private repos.
func NewClient() *Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	return &Client{gh: &RealGHCLI{}, fetcher: &HTTPFetcher{Token: token}}
}

// FetchUnityVersion reads ProjectSettings/ProjectVersion.txt from the repo's
// default branch. No gh CLI — plain HTTP either way:
//   - no token  → raw.githubusercontent.com anonymously (public repos).
//     NOTE: a token must NOT be sent here — raw.githubusercontent rejects
//     fine-grained PATs and turns public fetches into 404s.
//   - token set → GitHub contents API (works for private repos and accepts
//     fine-grained PATs; Accept header makes it return the raw body).
func (c *Client) FetchUnityVersion(owner, repo string) (string, error) {
	var fileURL string
	if c.fetcher.Authenticated() {
		fileURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/ProjectSettings/ProjectVersion.txt", owner, repo)
	} else {
		fileURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/HEAD/ProjectSettings/ProjectVersion.txt", owner, repo)
	}
	content, err := c.fetcher.Fetch(fileURL)
	if err != nil {
		return "", fmt.Errorf("fetch ProjectVersion.txt: %w", err)
	}
	return ParseUnityVersion(content)
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
