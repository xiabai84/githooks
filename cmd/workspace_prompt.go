package cmd

import (
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func workspaceSelectTemplates() *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "➣ {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "➣ {{ .Name | red | cyan }}",
		Details:  config.DetailTmpl,
	}
}

func workspaceSearcher(workspaces []types.Workspace) func(string, int) bool {
	return func(input string, index int) bool {
		name := strings.Replace(strings.ToLower(workspaces[index].Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}
}
