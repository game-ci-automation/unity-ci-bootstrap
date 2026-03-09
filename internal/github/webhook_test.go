package github

import (
	"fmt"
	"strings"
	"testing"
)

// sequentialFakeGHCLI returns different outputs for successive calls.
type sequentialFakeGHCLI struct {
	calls   []fakeCall
	current int
}

type fakeCall struct {
	output string
	err    error
}

func (f *sequentialFakeGHCLI) Run(args ...string) (string, error) {
	if f.current >= len(f.calls) {
		return "", fmt.Errorf("unexpected call #%d: %v", f.current, args)
	}
	call := f.calls[f.current]
	f.current++
	return call.output, call.err
}

func TestRegisterWebhook_Success(t *testing.T) {
	fake := &sequentialFakeGHCLI{
		calls: []fakeCall{
			{output: "[]"}, // GET hooks — empty list
			{output: `{"id": 123, "active": true}`}, // POST create hook
		},
	}
	client := &Client{gh: fake}

	err := client.RegisterWebhook("owner", "repo", "https://func.azurewebsites.net", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.current != 2 {
		t.Errorf("expected 2 gh calls, got %d", fake.current)
	}
}

func TestRegisterWebhook_AlreadyExists(t *testing.T) {
	fake := &sequentialFakeGHCLI{
		calls: []fakeCall{
			// GET hooks — one hook already pointing to our URL
			{output: `[{"id": 99, "config": {"url": "https://func.azurewebsites.net/api/github-webhook"}}]`},
		},
	}
	client := &Client{gh: fake}

	err := client.RegisterWebhook("owner", "repo", "https://func.azurewebsites.net", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only call GET, not POST (webhook already exists)
	if fake.current != 1 {
		t.Errorf("expected 1 gh call (skip creation), got %d", fake.current)
	}
}

func TestRegisterWebhook_ListError(t *testing.T) {
	fake := &sequentialFakeGHCLI{
		calls: []fakeCall{
			{err: fmt.Errorf("gh: not logged in")},
		},
	}
	client := &Client{gh: fake}

	err := client.RegisterWebhook("owner", "repo", "https://func.azurewebsites.net", "secret123")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestRegisterWebhook_CreateError(t *testing.T) {
	fake := &sequentialFakeGHCLI{
		calls: []fakeCall{
			{output: "[]"},                                    // GET — empty
			{err: fmt.Errorf("gh: 422 Validation Failed")},   // POST — fail
		},
	}
	client := &Client{gh: fake}

	err := client.RegisterWebhook("owner", "repo", "https://func.azurewebsites.net", "secret123")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "422") {
		t.Errorf("expected error to contain '422', got: %v", err)
	}
}
