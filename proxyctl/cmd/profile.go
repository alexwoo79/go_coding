package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/profile"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "管理系统代理 profile",
	Long: `将常用的系统代理配置保存为 profile，并可一键应用到系统代理。

内置 profile "direct" 表示关闭系统代理（直连）。

示例：
  proxyctl profile save clash   # 保存当前系统代理为 profile "clash"
  proxyctl profile list         # 列出全部 profile
  proxyctl profile use clash    # 应用 profile 到系统代理
  proxyctl profile use direct   # 关闭系统代理（直连）
  proxyctl profile remove clash # 删除 profile`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出全部 profile",
	Args:  usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ps, err := profile.List()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if len(ps) == 0 {
			fmt.Fprintln(out, "暂无 profile，可用 proxyctl profile save <名称> 保存当前系统代理。")
			fmt.Fprintln(out, "  direct (内置)\t关闭系统代理（直连）")
			return nil
		}
		for _, p := range ps {
			fmt.Fprintf(out, "  %-16s %s\n", p.Name, profileSummary(&p))
		}
		fmt.Fprintln(out, "  direct (内置)     关闭系统代理（直连）")
		return nil
	},
}

var profileSaveCmd = &cobra.Command{
	Use:   "save <名称>",
	Short: "把当前系统代理保存为 profile",
	Args:  usageArgs(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "direct" {
			return usageErr(errors.New("direct 是内置 profile，不能保存或覆盖"))
		}
		if !profile.ValidName(name) {
			return usageErr(fmt.Errorf("无效的 profile 名称 %q：仅支持字母、数字、-、_、.", name))
		}
		info, err := proxy.Get()
		if err != nil {
			return err
		}

		p := profile.Profile{Name: name}
		if info.HTTPEnable && info.HTTPHost != "" {
			p.HTTP = &proxy.EndpointState{Enabled: true, Host: info.HTTPHost, Port: info.HTTPPort}
		}
		if info.HTTPSEnable && info.HTTPSHost != "" {
			p.HTTPS = &proxy.EndpointState{Enabled: true, Host: info.HTTPSHost, Port: info.HTTPSPort}
		}
		if info.SOCKSEnable && info.SOCKSHost != "" {
			p.SOCKS = &proxy.EndpointState{Enabled: true, Host: info.SOCKSHost, Port: info.SOCKSPort}
		}
		if info.AutoConfig && info.AutoConfigURL != "" {
			p.PAC = &proxy.PACState{Enabled: true, URL: info.AutoConfigURL}
		}

		if err := profile.Save(p); err != nil {
			return fmt.Errorf("保存 profile 失败: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已保存 profile %q：%s\n", name, profileSummary(&p))
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <名称>",
	Short: "应用 profile 到系统代理",
	Args:  usageArgs(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		out := cmd.OutOrStdout()
		if name == "direct" {
			if err := proxy.Clear(); err != nil {
				return fmt.Errorf("关闭系统代理失败: %w", err)
			}
			fmt.Fprintln(out, "已应用 profile direct：系统代理已关闭（直连）。")
			return nil
		}

		p, err := profile.Load(name)
		if err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("未找到 profile %q，可用 proxyctl profile list 查看", name)
		}
		if err := proxy.ApplyProfile(p.HTTP, p.HTTPS, p.SOCKS, p.PAC); err != nil {
			return fmt.Errorf("应用 profile %q 失败: %w", name, err)
		}
		fmt.Fprintf(out, "已应用 profile %q：%s\n", name, profileSummary(p))
		fmt.Fprintln(out, "如需同步 git 代理，请运行: proxyctl apply")
		return nil
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <名称>",
	Short: "删除 profile",
	Args:  usageArgs(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "direct" {
			return usageErr(errors.New("direct 是内置 profile，不能删除"))
		}
		if err := profile.Remove(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已删除 profile %q。\n", name)
		return nil
	},
}

// profileSummary 生成 profile 的端点摘要文本。
func profileSummary(p *profile.Profile) string {
	var parts []string
	if p.HTTP != nil && p.HTTP.Enabled {
		parts = append(parts, "HTTP "+hostPort(p.HTTP.Host, p.HTTP.Port))
	}
	if p.HTTPS != nil && p.HTTPS.Enabled {
		parts = append(parts, "HTTPS "+hostPort(p.HTTPS.Host, p.HTTPS.Port))
	}
	if p.SOCKS != nil && p.SOCKS.Enabled {
		parts = append(parts, "SOCKS "+hostPort(p.SOCKS.Host, p.SOCKS.Port))
	}
	if p.PAC != nil && p.PAC.Enabled {
		parts = append(parts, "PAC "+p.PAC.URL)
	}
	if len(parts) == 0 {
		return "全部关闭"
	}
	return strings.Join(parts, "; ")
}

// hostPort 拼接 host:port，端口为空时省略端口。
func hostPort(host, port string) string {
	if port == "" {
		return host
	}
	return host + ":" + port
}

func init() {
	profileCmd.AddCommand(profileListCmd, profileSaveCmd, profileUseCmd, profileRemoveCmd)
}
