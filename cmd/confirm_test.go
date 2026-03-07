package cmd

import "testing"

func TestNewConfirmPrompt_ConstructiveDefaultsToYes(t *testing.T) {
	prompt := newConfirmPrompt("Input was correct", true)

	if !prompt.IsConfirm {
		t.Error("expected IsConfirm to be true")
	}
	if prompt.Default != "y" {
		t.Errorf("expected Default to be 'y' for constructive action, got %q", prompt.Default)
	}
	if prompt.Label != "Input was correct" {
		t.Errorf("expected Label 'Input was correct', got %q", prompt.Label)
	}
}

func TestNewConfirmPrompt_DestructiveDefaultsToNo(t *testing.T) {
	prompt := newConfirmPrompt("Delete workspace", false)

	if !prompt.IsConfirm {
		t.Error("expected IsConfirm to be true")
	}
	if prompt.Default != "" {
		t.Errorf("expected Default to be empty for destructive action, got %q", prompt.Default)
	}
	if prompt.Label != "Delete workspace" {
		t.Errorf("expected Label 'Delete workspace', got %q", prompt.Label)
	}
}
