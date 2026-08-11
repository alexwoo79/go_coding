package cmd

import (
	"fmt"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
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
		fmt.Fprintln(out, proxyLine("PAC", info.AutoConfig, info.PACURL()))

		fmt.Fprintln(out)
		fmt.Fprintln(out, "=== Git 全局代理 ===")
		gitCfg := git.New()
		httpVal, err := gitCfg.Get("http.proxy")
		if err != nil {
			fmt.Fprintf(out, "  http.proxy  = 读取失败: %v\n", err)
		} else {
			fmt.Fprintf(out, "  http.proxy  = %s\n", displayValue(httpVal))
		}
		httpsVal, err := gitCfg.Get("https.proxy")
		if err != nil {
			fmt.Fprintf(out, "  https.proxy = 读取失败: %v\n", err)
		} else {
			fmt.Fprintf(out, "  https.proxy = %s\n", displayValue(httpsVal))
		}

		return nil
	},
}

// proxyLine 生成单行代理状态文本。
func proxyLine(name string, enabled bool, url string) string {
	if !enabled {
		return fmt.Sprintf("%-7s代理: 未启用", name)
	}
	if url != "" {
		return fmt.Sprintf("%-7s代理已启用: %s", name, url)
	}
	return fmt.Sprintf("%-7s代理: 已启用", name)
}
