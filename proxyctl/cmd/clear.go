package cmd

import (
	"errors"
	"fmt"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/state"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "关闭系统代理并清除 git 全局代理",
	Long: `关闭系统代理（所有网络服务的 HTTP/HTTPS/SOCKS/PAC），
并清除 git 全局代理设置。若之前执行过 proxyctl apply，
本次操作前的状态可通过 proxyctl restore 恢复。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := proxy.Clear(); err != nil {
			return fmt.Errorf("清除系统代理失败: %w", err)
		}

		var unsetErr error
		gitCfg := git.New()
		for _, key := range git.Keys {
			if err := gitCfg.Unset(key); err != nil {
				unsetErr = errors.Join(unsetErr, err)
			}
		}
		if unsetErr != nil {
			return fmt.Errorf("清除 git 全局代理失败: %w", unsetErr)
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "已关闭系统代理。")
		fmt.Fprintln(out, "已清除 git 全局代理。")

		if path, err := state.DefaultPath(); err == nil {
			if snap, lerr := state.Load(path); lerr == nil && snap != nil {
				fmt.Fprintf(out, "apply 前的状态快照保留在 %s，执行 proxyctl restore 可恢复。\n", path)
			}
		}
		return nil
	},
}
