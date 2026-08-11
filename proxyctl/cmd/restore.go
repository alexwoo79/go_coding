package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/state"
	"github.com/alexwoo79/go_coding/proxyctl/internal/toolproxy"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从状态快照恢复系统代理与 git 全局代理",
	Long: `将系统代理与 git 全局代理恢复到上次 proxyctl apply 之前的状态。
快照由 apply 自动保存；恢复后快照保留，可重复执行。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := state.DefaultPath()
		if err != nil {
			return err
		}
		snap, err := state.Load(path)
		if err != nil {
			return err
		}
		if snap == nil {
			return errors.New("未找到状态快照，请先执行 proxyctl apply")
		}

		if err := proxy.RestoreSystem(snap.System); err != nil {
			return fmt.Errorf("恢复系统代理失败: %w", err)
		}
		if err := git.New().Restore(snap.Git); err != nil {
			return fmt.Errorf("恢复 git 全局代理失败: %w", err)
		}
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "已从快照恢复系统代理与 git 全局代理。")
		if names, rerr := restoreToolConfigs(toolproxy.Supported()); rerr != nil {
			return fmt.Errorf("恢复开发工具代理配置失败: %w", rerr)
		} else if len(names) > 0 {
			fmt.Fprintf(out, "已恢复开发工具代理配置：%s。\n", strings.Join(names, ", "))
		}
		return nil
	},
}
