package github

import (
	"fmt"
	"testing"
)

func TestParseUnityVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "standard version",
			content: "m_EditorVersion: 2022.2.1f1\nm_EditorVersionWithRevision: 2022.2.1f1 (4fead5835099)\n",
			want:    "2022.2.1f1",
		},
		{
			name:    "Unity 6 version",
			content: "m_EditorVersion: 6000.0.23f1\nm_EditorVersionWithRevision: 6000.0.23f1 (abcdef123456)\n",
			want:    "6000.0.23f1",
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "missing version line",
			content: "some random content\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUnityVersion(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchUnityVersion(t *testing.T) {
	fake := &fakeGHCLI{
		output: `{"content":"bV9FZGl0b3JWZXJzaW9uOiAyMDIyLjIuMWYxCm1fRWRpdG9yVmVyc2lvbldpdGhSZXZpc2lvbjogMjAyMi4yLjFmMSAoNGZlYWQ1ODM1MDk5KQo="}`,
	}
	client := &Client{gh: fake}

	version, err := client.FetchUnityVersion("JindoKimKor", "UnityGame3D-TeamTopChicken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "2022.2.1f1" {
		t.Errorf("got %q, want %q", version, "2022.2.1f1")
	}
}

func TestFetchUnityVersion_GHError(t *testing.T) {
	fake := &fakeGHCLI{
		err: fmt.Errorf("gh: not logged in"),
	}
	client := &Client{gh: fake}

	_, err := client.FetchUnityVersion("owner", "repo")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

type fakeGHCLI struct {
	output string
	err    error
}

func (f *fakeGHCLI) Run(args ...string) (string, error) {
	return f.output, f.err
}

func TestParseRepoOwnerName(t *testing.T) {
	tests := []struct {
		name    string
		repoURL string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:    "https URL",
			repoURL: "https://github.com/JindoKimKor/UnityGame3D-TeamTopChicken",
			owner:   "JindoKimKor",
			repo:    "UnityGame3D-TeamTopChicken",
		},
		{
			name:    "https URL with .git",
			repoURL: "https://github.com/JindoKimKor/UnityGame3D-TeamTopChicken.git",
			owner:   "JindoKimKor",
			repo:    "UnityGame3D-TeamTopChicken",
		},
		{
			name:    "invalid URL",
			repoURL: "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoOwnerName(tt.repoURL)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if owner != tt.owner || repo != tt.repo {
				t.Errorf("got (%q, %q), want (%q, %q)", owner, repo, tt.owner, tt.repo)
			}
		})
	}
}
