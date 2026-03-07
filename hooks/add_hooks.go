package hooks

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

// ValidateJiraKeyRegex checks if the given Jira key pattern is a valid regex.
func ValidateJiraKeyRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid Jira key regex %q: %w", pattern, err)
	}
	return nil
}

func AddWorkspace(newWorkspace *types.Workspace) error {
	if err := ValidateJiraKeyRegex(newWorkspace.ProjectKeyRE); err != nil {
		return err
	}

	// Check if a workspace with the same folder already exists
	ghConfig, err := ReadGitHooksConfig()
	if err != nil {
		return err
	}

	for i, ws := range ghConfig.Workspaces {
		if ws.Folder == newWorkspace.Folder {
			return mergeWorkspace(&ghConfig, i, newWorkspace)
		}
	}

	// Check for duplicate workspace name (different folder, same name = gitconfig conflict)
	for _, ws := range ghConfig.Workspaces {
		if strings.EqualFold(ws.Name, newWorkspace.Name) {
			return fmt.Errorf("workspace %q already exists (folder: %s). Use 'githooks update' to modify it", ws.Name, ws.Folder)
		}
	}

	// Warn if folder doesn't exist (non-blocking)
	expandedFolder := newWorkspace.Folder
	if strings.HasPrefix(expandedFolder, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			expandedFolder = home + expandedFolder[1:]
		}
	}
	if _, err := os.Stat(expandedFolder); os.IsNotExist(err) {
		fmt.Println(promptui.IconWarn+"  Warning: folder", newWorkspace.Folder, "does not exist")
	}

	if err := persistConfigAsJSON(newWorkspace); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GithooksConfigPath, "(added workspace entry)")
	if err := createWorkspaceGitConfig(newWorkspace); err != nil {
		return err
	}
	if err := updateGitConfigFile(newWorkspace); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Added workspace", newWorkspace.Name)
	return nil
}

func mergeWorkspace(ghConfig *types.GitHookConfig, existingIdx int, newWorkspace *types.Workspace) error {
	existing := &ghConfig.Workspaces[existingIdx]
	mergedKeys := mergeJiraKeys(existing.ProjectKeyRE, newWorkspace.ProjectKeyRE)
	existing.ProjectKeyRE = mergedKeys

	if err := WriteGitHooksConfig(ghConfig); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GithooksConfigPath, "(merged Jira keys)")
	// Rewrite the existing workspace's gitconfig file with merged keys
	if err := createWorkspaceGitConfig(existing); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Merged Jira keys into workspace", existing.Name, "→", mergedKeys)
	return nil
}

func mergeJiraKeys(existing, additional string) string {
	// Strip outer parentheses to get individual keys
	existingKeys := stripParens(existing)
	additionalKeys := stripParens(additional)

	// Collect all unique keys
	seen := make(map[string]bool)
	var keys []string
	for _, k := range strings.Split(existingKeys, "|") {
		k = strings.TrimSpace(k)
		if k != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	for _, k := range strings.Split(additionalKeys, "|") {
		k = strings.TrimSpace(k)
		if k != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}

	if len(keys) == 1 {
		return keys[0]
	}
	return "(" + strings.Join(keys, "|") + ")"
}

func stripParens(s string) string {
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return s[1 : len(s)-1]
	}
	return s
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
	sanitized := sanitizeGitConfig(string(bContent))
	tmpl, err := template.New("simple-hook-config").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(viewHeader + sanitized + config.GitConfigPatch)
	if err != nil {
		return fmt.Errorf("parsing git config template: %w", err)
	}
	return tmpl.Execute(os.Stdout, workspace)
}

// sanitizeGitConfig masks sensitive values in gitconfig content before display.
func sanitizeGitConfig(content string) string {
	sensitiveKeys := []string{"token", "password", "secret", "credential"}
	var result []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		masked := false
		for _, key := range sensitiveKeys {
			if strings.HasPrefix(trimmed, key+" =") || strings.HasPrefix(trimmed, key+"=") {
				// Keep the key but mask the value
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					indent := line[:len(line)-len(trimmed)]
					result = append(result, indent+strings.TrimSpace(parts[0])+" = ********")
					masked = true
					break
				}
			}
		}
		if !masked {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
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
	_, existsErr := os.Stat(workspaceGitConfigPath)
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
	if existsErr != nil {
		fmt.Println(promptui.IconGood+"  Created ", workspaceGitConfigPath)
	} else {
		fmt.Println(promptui.IconGood+"  Modified", workspaceGitConfigPath, "(updated workspace config)")
	}
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
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GitConfigPath, "(added includeIf block)")
	return nil
}
