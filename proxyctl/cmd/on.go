package cmd

import (
	"fmt"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/toolproxy"
	"github.com/spf13/cobra"
)

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "自动检测代理程序并一键开启",
	Long: `自动检测系统中运行中的代理程序（Clash/Mihomo/v2ray/xray/sing-box 等）
及其端口，然后把代理应用到：
  1. 系统代理（macOS 所有网络服务的 HTTP/SOCKS）
  2. git 全局代理（http.proxy / https.proxy）
  3. 开发工具配置（npm / pnpm / pip / cargo / docker / brew）
  4. 终端环境变量（新开终端自动生效，或手动 eval）
执行前会保存快照，可用 proxyctl off / restore 恢复。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := proxy.Detect()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "检测到代理程序: %s (%s)\n", d.Command, d.URL())

		// 1. 先保存快照（系统 + git + tools），保证 off / restore 可恢复
		path, err := saveStateSnapshot()
		if err != nil {
			return err
		}
		if err := snapshotToolConfigs(toolproxy.Supported()); err != nil {
			return err
		}

		// 2. 设置系统代理
		var httpEP, socksEP *proxy.EndpointState
		if d.Scheme == "socks5h" {
			socksEP = &proxy.EndpointState{Enabled: true, Host: d.Host, Port: d.Port}
		} else {
			httpEP = &proxy.EndpointState{Enabled: true, Host: d.Host, Port: d.Port}
		}
		if err := proxy.ApplyProfile(httpEP, nil, socksEP, nil); err != nil {
			return fmt.Errorf("设置系统代理失败: %w", err)
		}
		fmt.Fprintf(out, "已设置系统代理: %s\n", d.URL())

		// 3. 设置 git 代理
		url := d.URL()
		gitCfg := git.New()
		if err := gitCfg.Set("http.proxy", url); err != nil {
			return fmt.Errorf("设置 git http.proxy 失败: %w", err)
		}
		if err := gitCfg.Set("https.proxy", url); err != nil {
			return fmt.Errorf("设置 git https.proxy 失败: %w", err)
		}
		fmt.Fprintf(out, "已设置 git 代理: %s\n", url)

		// 4. 应用开发工具配置
		if err := toolproxy.ApplyTo(toolproxy.Supported(), url); err != nil {
			return fmt.Errorf("应用开发工具代理失败: %w", err)
		}
		fmt.Fprintf(out, "已应用开发工具代理：%s\n", strings.Join(toolNames(toolproxy.Supported()), ", "))

		// 5. 提示终端环境变量
		fmt.Fprintf(out, "已保存快照: %s\n", path)
		fmt.Fprintln(out, "当前终端如需立即生效，请执行: eval \"$(proxyctl env)\"")
		fmt.Fprintln(out, "新开终端自动生效可先运行: proxyctl env install")
		return nil
	},
}

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "一键直连（等价于 clear）",
	Long: `关闭系统代理（含 WPAD）、清除 git 与开发工具（npm/pnpm/pip/cargo/docker/brew）
的代理配置，并提示当前终端执行 eval "$(proxyctl env --clear)" 立即直连。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClear(cmd)
	},
}
