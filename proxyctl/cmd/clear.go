package cmd

import (
	"fmt"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "关闭系统代理并清除 git 全局代理",
	Long: `关闭系统代理（所有网络服务的 HTTP/HTTPS/SOCKS/PAC），
并清除 git 全局代理设置。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := proxy.Clear(); err != nil {
			return fmt.Errorf("清除系统代理失败: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "已关闭系统代理。")

		// 配置项不存在时 --unset 会报错，忽略即可
		_ = gitUnset("http.proxy")
		_ = gitUnset("https.proxy")
		fmt.Fprintln(cmd.OutOrStdout(), "已清除 git 全局代理。")
		return nil
	},
}
