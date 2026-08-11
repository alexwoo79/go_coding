package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/toolproxy"
	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "管理开发工具（npm/cargo/pip/docker）的代理配置",
	Long: `读写各开发工具自己的代理配置文件，配合 env（终端环境变量）覆盖所有 CLI 工具：
  npm     ~/.npmrc                proxy / https-proxy
  pip     ~/.config/pip/pip.conf  [global] proxy
  cargo   ~/.cargo/config.toml    [http] proxy
  docker  ~/.docker/config.json   proxies.default.*
环境变量型工具（brew/rustup/uv 等）由 proxyctl env 覆盖。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var toolsListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看各工具当前的代理配置",
	Args:  usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		tools := toolproxy.Supported()
		all, err := toolproxy.ReadAll(tools)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		for _, t := range tools {
			fmt.Fprintf(out, "%-8s %s\n", t.Name(), t.Path())
			vals := all[t.Name()]
			if len(vals) == 0 {
				fmt.Fprintln(out, "          未配置代理")
				continue
			}
			for _, k := range sortedKeys(vals) {
				if v := vals[k]; v != nil {
					fmt.Fprintf(out, "          %s = %s\n", k, *v)
				}
			}
		}
		return nil
	},
}

var toolsApplyCmd = &cobra.Command{
	Use:   "apply [工具...]",
	Short: "把当前系统代理写入工具的配置文件",
	Long: `把当前系统代理（HTTP 优先，SOCKS 回退）写入指定工具（默认全部）的配置文件。
写入前会保存快照，可用 proxyctl tools restore 恢复。`,
	Args: usageArgs(cobra.ArbitraryArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		tools, err := toolproxy.Select(args)
		if err != nil {
			return usageErr(err)
		}
		proxyURL, err := systemProxyURL()
		if err != nil {
			return err
		}
		if err := snapshotToolConfigs(tools); err != nil {
			return err
		}
		if err := toolproxy.ApplyTo(tools, proxyURL); err != nil {
			return fmt.Errorf("应用工具代理失败: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已为 %s 写入代理: %s\n", strings.Join(toolNames(tools), ", "), proxyURL)
		return nil
	},
}

var toolsClearCmd = &cobra.Command{
	Use:   "clear [工具...]",
	Short: "清除工具的代理配置",
	Args:  usageArgs(cobra.ArbitraryArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		tools, err := toolproxy.Select(args)
		if err != nil {
			return usageErr(err)
		}
		if err := snapshotToolConfigs(tools); err != nil {
			return err
		}
		if err := toolproxy.Clear(tools); err != nil {
			return fmt.Errorf("清除工具代理失败: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已清除 %s 的代理配置。\n", strings.Join(toolNames(tools), ", "))
		return nil
	},
}

var toolsRestoreCmd = &cobra.Command{
	Use:   "restore [工具...]",
	Short: "从快照恢复工具的代理配置",
	Args:  usageArgs(cobra.ArbitraryArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		tools, err := toolproxy.Select(args)
		if err != nil {
			return usageErr(err)
		}
		names, err := restoreToolConfigs(tools)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			return errors.New("未找到工具代理快照，请先运行 proxyctl tools apply")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已恢复 %s 的代理配置。\n", strings.Join(names, ", "))
		return nil
	},
}

// systemProxyURL 返回当前系统代理 URL（HTTP 优先，SOCKS 回退）。
func systemProxyURL() (string, error) {
	info, err := proxy.Get()
	if err != nil {
		return "", err
	}
	u := info.HTTPProxyURL()
	if u == "" {
		u = info.SOCKSPProxyURL()
	}
	if u == "" {
		return "", errors.New("未检测到可用的系统代理，无法应用；请先启用代理或使用 proxyctl profile use")
	}
	return u, nil
}

// snapshotToolConfigs 保存指定工具的当前配置到快照（覆盖式）。
func snapshotToolConfigs(tools []toolproxy.Tool) error {
	path, err := toolproxy.DefaultStatePath()
	if err != nil {
		return err
	}
	snap, err := toolproxy.LoadState(path)
	if err != nil {
		return err
	}
	if snap == nil {
		snap = &toolproxy.Snapshot{Version: toolproxy.StateVersion}
	}
	if snap.Tools == nil {
		snap.Tools = map[string]map[string]*string{}
	}
	all, err := toolproxy.ReadAll(tools)
	if err != nil {
		return err
	}
	for name, vals := range all {
		snap.Tools[name] = vals
	}
	snap.Timestamp = time.Now()
	if err := toolproxy.SaveState(path, snap); err != nil {
		return fmt.Errorf("保存工具快照失败: %w", err)
	}
	return nil
}

// restoreToolConfigs 恢复指定工具；返回实际恢复的工具名。
func restoreToolConfigs(tools []toolproxy.Tool) ([]string, error) {
	path, err := toolproxy.DefaultStatePath()
	if err != nil {
		return nil, err
	}
	snap, err := toolproxy.LoadState(path)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return nil, nil
	}
	var names []string
	var selected []toolproxy.Tool
	for _, t := range tools {
		if _, ok := snap.Tools[t.Name()]; ok {
			names = append(names, t.Name())
			selected = append(selected, t)
		}
	}
	if len(selected) == 0 {
		return nil, nil
	}
	if err := toolproxy.Restore(selected, snap.Tools); err != nil {
		return nil, fmt.Errorf("恢复工具代理失败: %w", err)
	}
	return names, nil
}

// clearToolConfigs 清除全部工具代理配置；无快照时先保存（保证 restore 可用）。
func clearToolConfigs() ([]string, error) {
	tools := toolproxy.Supported()
	path, err := toolproxy.DefaultStatePath()
	if err != nil {
		return nil, err
	}
	snap, err := toolproxy.LoadState(path)
	if err != nil {
		return nil, err
	}
	var missing []toolproxy.Tool
	for _, t := range tools {
		if snap == nil || snap.Tools[t.Name()] == nil {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		if err := snapshotToolConfigs(missing); err != nil {
			return nil, err
		}
	}
	if err := toolproxy.Clear(tools); err != nil {
		return nil, err
	}
	return toolNames(tools), nil
}

// toolNames 返回排序后的工具名列表。
func toolNames(tools []toolproxy.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name())
	}
	sort.Strings(names)
	return names
}

func sortedKeys(m map[string]*string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func init() {
	toolsCmd.AddCommand(toolsListCmd, toolsApplyCmd, toolsClearCmd, toolsRestoreCmd)
}
