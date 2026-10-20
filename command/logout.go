package command

import (
	"github.com/spf13/cobra"

	"modstart-cli/lib/api"
	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "清除本地凭证",
	RunE: func(cmd *cobra.Command, args []string) error {
		site, err := activeSite()
		if err != nil {
			return err
		}
		client := api.New(site)
		if err := client.Logout(); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		console.PrintJSON(map[string]interface{}{
			"ok":       true,
			"adminUrl": site.AdminURL,
		})
		return nil
	},
}
