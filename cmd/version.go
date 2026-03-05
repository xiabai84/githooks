package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefan-niemeyer/githooks/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Prints the githooks version",
	Long:    `Prints the githooks version`,
	Example: `githooks version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildinfo.GetBuildInfo().String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
