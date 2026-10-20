package command

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"modstart-cli/lib/api"
	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var (
	flagToken    string
	flagName     string
	flagAdminURL string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "校验凭证并保存（无状态，不使用 Cookie）",
	RunE: func(cmd *cobra.Command, args []string) error {
		adminURL := config.NormalizeURL(flagAdminURL)
		if adminURL == "" {
			adminURL = config.NormalizeURL(console.Prompt("后台管理地址(如 https://demo.com/admin): "))
		}
		if adminURL == "" {
			return fmt.Errorf("后台管理地址不能为空")
		}

		token := flagToken
		if token == "" {
			token = console.Prompt("CLI 凭证: ")
		}
		if token == "" {
			return fmt.Errorf("CLI 凭证不能为空")
		}

		name := flagName
		if name == "" {
			name, _ = os.Hostname()
		}

		siteURL, adminPath := splitAdminURL(adminURL)
		site := &config.Site{
			URL:       siteURL,
			AdminURL:  adminURL,
			App:       "admin",
			AdminPath: adminPath,
			Token:     token,
			Name:      name,
		}

		client := api.New(site)
		resp, err := client.Login(token, name)
		if err != nil {
			return err
		}

		cfg.Put(site)
		if err := config.Save(cfg); err != nil {
			return err
		}

		var loginData interface{}
		if len(resp.Data) > 0 {
			_ = json.Unmarshal(resp.Data, &loginData)
		}
		console.PrintJSON(map[string]interface{}{
			"ok":       true,
			"adminUrl": site.AdminURL,
			"name":     name,
			"login":    loginData,
		})
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&flagAdminURL, "admin-url", "", "后台管理地址（如 https://demo.com/admin）")
	loginCmd.Flags().StringVar(&flagToken, "token", "", "后台生成的 CLI 凭证")
	loginCmd.Flags().StringVar(&flagName, "name", "", "设备名称，默认取本机 hostname")
}

// splitAdminURL 拆分后台管理地址为「站点根地址 + 后台路径」
func splitAdminURL(rawURL string) (string, string) {
	rawURL = config.NormalizeURL(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "admin"
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return strings.TrimRight(u.String(), "/"), "admin"
	}
	segs := strings.Split(path, "/")
	adminPath := segs[len(segs)-1]
	segs = segs[:len(segs)-1]
	u.Path = "/" + strings.Join(segs, "/")
	if u.Path == "/" {
		u.Path = ""
	}
	return strings.TrimRight(u.String(), "/"), adminPath
}
