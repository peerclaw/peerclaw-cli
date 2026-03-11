package cmd

import "testing"

func TestRunInvoke_NoArgs(t *testing.T) {
	code := RunInvoke([]string{}, "http://localhost:8080")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunInvoke_NoMessage(t *testing.T) {
	code := RunInvoke([]string{"agent-123"}, "http://localhost:8080")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
