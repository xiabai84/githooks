package hooks

import (
	"os"
	"strings"

	"github.com/xiabai84/githooks/types"
)

func GetWorkspaceIndex(workspaces []types.Workspace) (int, error) {
	var matchIdx int

	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	if len(cwd) == 0 {
		return matchIdx, nil
	}
	cwd += "/"

	longestMatch := 0
	for idx, workspace := range workspaces {
		wsFolder := strings.Replace(workspace.Folder, "~", home, 1)
		if strings.HasPrefix(cwd, wsFolder) && len(wsFolder) > longestMatch {
			longestMatch = len(wsFolder)
			matchIdx = idx
		}
	}

	return matchIdx, nil
}
