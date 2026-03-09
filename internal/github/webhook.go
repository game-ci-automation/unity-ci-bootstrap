package github

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RegisterWebhook creates a GitHub webhook for push events on the given repo.
// If a webhook with the same URL already exists, it skips creation.
func (c *Client) RegisterWebhook(owner, repo, functionAppURL, secret string) error {
	hookURL := strings.TrimRight(functionAppURL, "/") + "/api/github-webhook"

	// Check for existing webhooks
	endpoint := fmt.Sprintf("repos/%s/%s/hooks", owner, repo)
	output, err := c.gh.Run("api", endpoint)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	var hooks []struct {
		Config struct {
			URL string `json:"url"`
		} `json:"config"`
	}
	if err := json.Unmarshal([]byte(output), &hooks); err != nil {
		return fmt.Errorf("failed to parse webhooks: %w", err)
	}

	for _, h := range hooks {
		if h.Config.URL == hookURL {
			return nil // already exists
		}
	}

	// Create webhook
	_, err = c.gh.Run("api", endpoint, "--method", "POST",
		"-f", fmt.Sprintf("config[url]=%s", hookURL),
		"-f", "config[content_type]=json",
		"-f", fmt.Sprintf("config[secret]=%s", secret),
		"-f", "events[]=push",
		"-F", "active=true",
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}
