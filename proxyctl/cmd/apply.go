package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/state"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "根据系统代理自动设置 git 全局代理",
	Long: `根据当前 macOS 系统代理设置 git 全局代理（http.proxy / https.proxy），
并提示在当前终端手动执行的 export 命令（子进程无法影响父 shell 环境变量）。
执行前会保存系统代理与 git 代理状态快照，之后可用 proxyctl restore 恢复。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := proxy.Get()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		// 优先使用 HTTP 代理，其次 SOCKS 代理
		url := info.HTTPProxyURL()
		kind, host, port := "HTTP", info.HTTPHost, info.HTTPPort
		if url == "" {
			url = info.SOCKSPProxyURL()
			kind, host, port = "SOCKS", info.SOCKSHost, info.SOCKSPort
		}

		if url == "" {
			fmt.Fprintln(out, "未检测到可用的系统代理，网络将使用直连模式。")
			fmt.Fprintln(out, "如需清除 git 代理，请执行: proxyctl clear")
			return nil
		}

		// 先保存快照再修改配置：clear 之后可用 restore 恢复 apply 前的状态。
		path, err := saveStateSnapshot()
		if err != nil {
			return err
		}
		gitCfg := git.New()

		fmt.Fprintf(out, "检测到 %s 代理: %s:%s\n", kind, host, port)
		fmt.Fprintf(out, "已保存状态快照: %s\n", path)
		if err := gitCfg.Set("http.proxy", url); err != nil {
			return fmt.Errorf("设置 git http.proxy 失败: %w", err)
		}
		if err := gitCfg.Set("https.proxy", url); err != nil {
			return fmt.Errorf("设置 git https.proxy 失败: %w", err)
		}
		fmt.Fprintln(out, "已为 git 设置代理：")
		fmt.Fprintf(out, "  http.proxy  = %s\n", url)
		fmt.Fprintf(out, "  https.proxy = %s\n", url)
		printExportHint(out, url)
		return nil
	},
}

// saveStateSnapshot 保存当前系统代理与 git 代理快照，返回快照路径。
func saveStateSnapshot() (string, error) {
	systemSnap, err := proxy.SnapshotSystem()
	if err != nil {
		return "", fmt.Errorf("保存系统代理快照失败（已中止，避免覆盖原配置）: %w", err)
	}
	gitSnap, err := git.New().Snapshot()
	if err != nil {
		return "", fmt.Errorf("保存 git 代理快照失败（已中止，避免覆盖原配置）: %w", err)
	}
	path, err := state.DefaultPath()
	if err != nil {
		return "", err
	}
	if err := state.Save(path, &state.State{
		Version:   state.Version,
		Timestamp: time.Now(),
		System:    systemSnap,
		Git:       gitSnap,
	}); err != nil {
		return "", fmt.Errorf("保存状态快照失败（已中止，避免覆盖原配置）: %w", err)
	}
	return path, nil
}

// printExportHint 输出在当前 shell 手动设置代理环境变量的提示。
func printExportHint(w io.Writer, url string) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "提示：如需在当前终端生效（子进程无法影响父 shell），请手动执行：")
	fmt.Fprintf(w, "  export http_proxy=\"%s\"\n", url)
	fmt.Fprintf(w, "  export https_proxy=\"%s\"\n", url)
}
