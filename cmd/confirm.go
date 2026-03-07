package cmd

import "github.com/manifoldco/promptui"

// newConfirmPrompt creates a confirmation prompt.
// Constructive actions (add, update) default to Yes.
// Destructive actions (delete, uninstall) default to No.
func newConfirmPrompt(label string, defaultYes bool) promptui.Prompt {
	p := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	if defaultYes {
		p.Default = "y"
	}
	return p
}
