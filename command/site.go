package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "管理已登记的站点",
}

var siteListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已登记站点",
	RunE: func(cmd *cobra.Command, args []string) error {
		sites := make([]map[string]interface{}, 0, len(cfg.Sites))
		for _, s := range cfg.Sites {
			adminURL := s.AdminURL
			if adminURL == "" {
				adminURL = s.URL
			}
			sites = append(sites, map[string]interface{}{
				"url":      s.URL,
				"adminUrl": adminURL,
				"name":     s.Name,
				"current":  s.URL == cfg.Current,
			})
		}
		console.PrintJSON(map[string]interface{}{
			"current": cfg.Current,
			"sites":   sites,
		})
		return nil
	},
}

var siteUseCmd = &cobra.Command{
	Use:   "use <后台管理地址>",
	Short: "切换当前站点",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := siteKey(args[0])
		if _, ok := cfg.Sites[key]; !ok {
			return fmt.Errorf("站点未登记: %s", key)
		}
		cfg.Current = key
		if err := config.Save(cfg); err != nil {
			return err
		}
		console.PrintJSON(map[string]interface{}{
			"ok":      true,
			"current": key,
		})
		return nil
	},
}

var siteRemoveCmd = &cobra.Command{
	Use:   "remove <后台管理地址>",
	Short: "移除站点及其凭证",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := siteKey(args[0])
		if err := cfg.Remove(key); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		console.PrintJSON(map[string]interface{}{
			"ok":     true,
			"remove": key,
		})
		return nil
	},
}

func init() {
	siteCmd.AddCommand(siteListCmd, siteUseCmd, siteRemoveCmd)
}

// siteKey 将后台管理地址转换为站点根地址（配置键）
func siteKey(raw string) string {
	siteURL, _ := splitAdminURL(raw)
	return siteURL
}
