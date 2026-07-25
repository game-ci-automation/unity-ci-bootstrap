package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	fake := &fakeFetcher{
		content: "m_EditorVersion: 2022.2.1f1\nm_EditorVersionWithRevision: 2022.2.1f1 (4fead5835099)\n",
	}
	client := &Client{fetcher: fake}

	version, err := client.FetchUnityVersion("JindoKimKor", "UnityGame3D-TeamTopChicken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "2022.2.1f1" {
		t.Errorf("got %q, want %q", version, "2022.2.1f1")
	}
	wantURL := "https://raw.githubusercontent.com/JindoKimKor/UnityGame3D-TeamTopChicken/HEAD/ProjectSettings/ProjectVersion.txt"
	if fake.gotURL != wantURL {
		t.Errorf("fetched %q, want %q", fake.gotURL, wantURL)
	}
}

func TestFetchUnityVersion_FetchError(t *testing.T) {
	fake := &fakeFetcher{
		err: fmt.Errorf("GET failed: 404 Not Found"),
	}
	client := &Client{fetcher: fake}

	_, err := client.FetchUnityVersion("owner", "repo")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

type fakeFetcher struct {
	content       string
	err           error
	gotURL        string
	authenticated bool
}

func (f *fakeFetcher) Fetch(url string) (string, error) {
	f.gotURL = url
	return f.content, f.err
}

func (f *fakeFetcher) Authenticated() bool {
	return f.authenticated
}

func TestFetchUnityVersion_AuthenticatedUsesAPI(t *testing.T) {
	fake := &fakeFetcher{
		content:       "m_EditorVersion: 2022.2.1f1\n",
		authenticated: true,
	}
	client := &Client{fetcher: fake}

	if _, err := client.FetchUnityVersion("owner", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantURL := "https://api.github.com/repos/owner/repo/contents/ProjectSettings/ProjectVersion.txt"
	if fake.gotURL != wantURL {
		t.Errorf("fetched %q, want %q", fake.gotURL, wantURL)
	}
}

func TestHTTPFetcher_TokenHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		fmt.Fprint(w, "m_EditorVersion: 2022.2.1f1\n")
	}))
	defer srv.Close()

	// With token → Authorization header set (private repo support)
	f := &HTTPFetcher{Token: "pat-123"}
	if _, err := f.Fetch(srv.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "token pat-123" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "token pat-123")
	}

	// Without token → no header (public repo)
	f = &HTTPFetcher{}
	if _, err := f.Fetch(srv.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization = %q, want empty", gotAuth)
	}
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
