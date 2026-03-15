package cmd

import (
	"strings"
	"testing"
)

func TestRunCompletionNoArgs(t *testing.T) {
	if code := RunCompletion(nil); code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
}

func TestRunCompletionUnknownShell(t *testing.T) {
	if code := RunCompletion([]string{"powershell"}); code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
}

func TestBashCompletionContainsPeerclaw(t *testing.T) {
	if !strings.Contains(bashCompletion, "complete -F _peerclaw peerclaw") {
		t.Error("bash completion missing complete command")
	}
}

func TestZshCompletionContainsCompdef(t *testing.T) {
	if !strings.Contains(zshCompletion, "compdef _peerclaw peerclaw") {
		t.Error("zsh completion missing compdef")
	}
}

func TestFishCompletionContainsCommands(t *testing.T) {
	for _, cmd := range []string{"agent", "invoke", "send", "health", "completion"} {
		if !strings.Contains(fishCompletion, cmd) {
			t.Errorf("fish completion missing command %q", cmd)
		}
	}
}
