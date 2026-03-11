package cmd

import "testing"

func TestRunInbox_NoArgs(t *testing.T) {
	code := RunInbox([]string{}, "http://localhost:8080")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunInbox_UnknownSubcommand(t *testing.T) {
	code := RunInbox([]string{"unknown"}, "http://localhost:8080")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
