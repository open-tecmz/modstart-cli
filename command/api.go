package command

import (
	"github.com/spf13/cobra"

	"modstart-cli/lib/api"
	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "查看站点可用的后台接口",
}

var apiListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可用的后台接口",
	RunE: func(cmd *cobra.Command, args []string) error {
		site, err := activeSite()
		if err != nil {
			return err
		}
		client := api.New(site)
		list, err := client.ApiList()
		if err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		if flagJSON {
			console.PrintJSON(list)
			return nil
		}
		console.PrintApis(list.Apis)
		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiListCmd)
}
