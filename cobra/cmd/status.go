package cmd

import (
	"fmt"

	"github.com/alexwoo79/go_coding/cobra/internal/proxy"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前 macOS 系统代理与 git 代理状态",
	Args:  usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := proxy.Get()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		fmt.Fprintln(out, "=== macOS 系统代理 ===")
		fmt.Fprintln(out, proxyLine("HTTP", info.HTTPEnable, info.HTTPProxyURL()))
		fmt.Fprintln(out, proxyLine("HTTPS", info.HTTPSEnable, info.HTTPSProxyURL()))
		fmt.Fprintln(out, proxyLine("SOCKS", info.SOCKSEnable, info.SOCKSPProxyURL()))
		fmt.Fprintf(out, "%-7s自动代理: %s\n", "PAC", enabledText(info.AutoConfig))

		fmt.Fprintln(out)
		fmt.Fprintln(out, "=== Git 全局代理 ===")
		fmt.Fprintf(out, "  http.proxy  = %s\n", displayValue(gitGet("http.proxy")))
		fmt.Fprintf(out, "  https.proxy = %s\n", displayValue(gitGet("https.proxy")))

		return nil
	},
}

// proxyLine 生成单行代理状态文本。
func proxyLine(name string, enabled bool, url string) string {
	if enabled && url != "" {
		return fmt.Sprintf("%-7s代理已启用: %s", name, url)
	}
	return fmt.Sprintf("%-7s代理: 未启用", name)
}

// enabledText 将布尔值转换为中文状态文本。
func enabledText(v bool) string {
	if v {
		return "已启用"
	}
	return "未启用"
}
