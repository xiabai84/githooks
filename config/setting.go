package config

import (
	"log"
	"os"
	"path/filepath"
)

const GitHooksFolder = ".githooks"
const GitHooksConfigFolder = "config"
const GithooksLogName = "githooks.log"
const GithooksConfigName = "githooks.json"
const CommitMsgName = "commit-msg"
const GitConfigFilename = ".gitconfig"
const GitHooksConfigPrefix = "gitconfig"

const ConfigFilePermission os.FileMode = 0644
const ExecutableFilePermission os.FileMode = 0755

type Paths struct {
	HomeDir            string
	HookDir            string
	HookConfigDir      string
	GithooksLogPath    string
	GithooksConfigPath string
	CommitMsgPath      string
	GitConfigPath      string
}

var Default Paths

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to determine home directory: %v", err)
	}
	Default = NewPaths(homeDir)
}

func NewPaths(homeDir string) Paths {
	hookDir := filepath.Join(homeDir, GitHooksFolder)
	hookConfigDir := filepath.Join(hookDir, GitHooksConfigFolder)
	return Paths{
		HomeDir:            homeDir,
		HookDir:            hookDir,
		HookConfigDir:      hookConfigDir,
		GithooksLogPath:    filepath.Join(hookConfigDir, GithooksLogName),
		GithooksConfigPath: filepath.Join(hookConfigDir, GithooksConfigName),
		CommitMsgPath:      filepath.Join(hookDir, CommitMsgName),
		GitConfigPath:      filepath.Join(homeDir, GitConfigFilename),
	}
}

var GitConfigPatch = `[includeIf "gitdir:{{ .Folder }}"]
    path = ` + GitHooksFolder + `/` + GitHooksConfigFolder + `/` + GitHooksConfigPrefix + `-{{ toLower .Name }}
`

var HooksConfigTmpl = `[core]
    hooksPath=~/` + GitHooksFolder + `
[user]
    jiraProjects={{ .ProjectKeyRE }}
`

var DetailTmpl = `
{{ if ne .Name "Quit" }}
------------------ Workspace Configuration --------------------
Name: {{ .Name | faint }}
Folder: {{ .Folder | faint }}
Jira Project Key RegEx: {{ .ProjectKeyRE | faint }}
{{ end }}
`
