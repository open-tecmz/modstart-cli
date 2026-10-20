package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"modstart-cli/lib/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("modstart-cli %s (%s)\n", version.VERSION, version.BUILD_TIME)
	},
}
