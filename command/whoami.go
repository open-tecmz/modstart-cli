package command

import (
	"github.com/spf13/cobra"

	"modstart-cli/lib/api"
	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "显示当前登录的管理员",
	RunE: func(cmd *cobra.Command, args []string) error {
		site, err := activeSite()
		if err != nil {
			return err
		}
		client := api.New(site)
		me, err := client.Me()
		if err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		console.PrintJSON(me)
		return nil
	},
}
