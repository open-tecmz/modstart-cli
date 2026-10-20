package command

import (
	"github.com/spf13/cobra"

	"modstart-cli/lib/config"
)

var (
	flagSite string
	flagJSON bool
	cfg      *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "modstart-cli",
	Short: "ModStart CLI 后台管理工具",
	Long: `ModStart CLI 通过后台生成的凭证登录站点，直接调用站点后台已有接口。

使用流程：
  1. 后台「系统设置 → 接口服务 → CLI 授权管理」生成凭证
  2. modstart-cli login --admin-url=https://your-domain.com/admin --token=<凭证>
  3. modstart-cli api list / modstart-cli call <key> '{"page":1}'

所有命令统一输出 JSON（中文不转义）。`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "version" {
			return nil
		}
		var err error
		cfg, err = config.Load()
		return err
	},
}

// Execute 执行命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagSite, "site", "", "站点根地址，默认使用当前站点")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "以 JSON 输出（api list 默认文字）")

	rootCmd.AddCommand(
		loginCmd,
		logoutCmd,
		whoamiCmd,
		siteCmd,
		apiCmd,
		callCmd,
		versionCmd,
	)
}

// activeSite 当前操作站点
func activeSite() (*config.Site, error) {
	return cfg.Find(flagSite)
}
