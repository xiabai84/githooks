package hooks

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func AddWorkspace(newWorkspace *types.Workspace) error {
	if err := persistConfigAsJSON(newWorkspace); err != nil {
		return err
	}
	if err := createWorkspaceGitConfig(newWorkspace); err != nil {
		return err
	}
	return updateGitConfigFile(newWorkspace)
}

func PreviewConfig(newWorkspace *types.Workspace) error {
	if err := previewGitConfigFile(newWorkspace); err != nil {
		return err
	}
	return previewWorkspaceGitConfig(newWorkspace)
}

func CheckConfigFiles() error {
	requiredPaths := []string{
		config.Default.GitConfigPath,
		config.Default.HookDir,
		config.Default.HookConfigDir,
		config.Default.CommitMsgPath,
		config.Default.GithooksConfigPath,
	}
	for _, p := range requiredPaths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("%s doesn't exist, please execute 'githooks init' first", p)
		}
	}
	return nil
}

func previewGitConfigFile(workspace *types.Workspace) error {
	viewHeader := "========================== ~/" + config.GitConfigFilename + " ==========================\n"
	bContent, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		return fmt.Errorf("git configuration file %s doesn't exist, please setup git first", config.Default.GitConfigPath)
	}
	tmpl, err := template.New("simple-hook-config").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(viewHeader + string(bContent) + config.GitConfigPatch)
	if err != nil {
		return fmt.Errorf("parsing git config template: %w", err)
	}
	return tmpl.Execute(os.Stdout, workspace)
}

func previewWorkspaceGitConfig(workspace *types.Workspace) error {
	viewHeader := "========================== ~/" + config.GitHooksFolder + "/" + config.GitHooksConfigFolder + "/" + config.GitHooksConfigPrefix + "-" + strings.ToLower(workspace.Name) + " ==========================\n"
	tmpl, err := template.New("simple-jira-config").Parse(viewHeader + config.HooksConfigTmpl)
	if err != nil {
		return fmt.Errorf("parsing workspace config template: %w", err)
	}
	return tmpl.Execute(os.Stdout, workspace)
}

func persistConfigAsJSON(workspace *types.Workspace) error {
	ghConfig, err := ReadGitHooksConfig()
	if err != nil {
		return err
	}
	ghConfig.Workspaces = append(ghConfig.Workspaces, *workspace)
	return WriteGitHooksConfig(&ghConfig)
}

func createWorkspaceGitConfig(workspace *types.Workspace) error {
	workspaceGitConfigPath := config.Default.HookConfigDir + "/" + config.GitHooksConfigPrefix + "-" + strings.ToLower(workspace.Name)
	tmpl, err := template.New("jira-config").Parse(config.HooksConfigTmpl)
	if err != nil {
		return fmt.Errorf("parsing jira config template: %w", err)
	}
	f, err := os.Create(workspaceGitConfigPath)
	if err != nil {
		return fmt.Errorf("creating workspace git config: %w", err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, workspace); err != nil {
		return fmt.Errorf("executing jira config template: %w", err)
	}
	fmt.Println(promptui.IconGood+"  Create new file:", workspaceGitConfigPath)
	return nil
}

func updateGitConfigFile(workspace *types.Workspace) error {
	bContent, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		return fmt.Errorf("git configuration file %s doesn't exist, please setup this first", config.Default.GitConfigPath)
	}
	tmpl, err := template.New("simple-hook-config").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(string(bContent) + config.GitConfigPatch)
	if err != nil {
		return fmt.Errorf("parsing git config template: %w", err)
	}
	f, err := os.Create(config.Default.GitConfigPath)
	if err != nil {
		return fmt.Errorf("creating git config: %w", err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, workspace); err != nil {
		return fmt.Errorf("executing git config template: %w", err)
	}
	fmt.Println(promptui.IconGood+"  Updated file:", config.Default.GitConfigPath)
	return nil
}
