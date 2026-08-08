package cmd

import (
	"errors"
	"fmt"
	"text/tabwriter"

	portutil "github.com/alexwoo79/go_coding/cobra/internal/port"
	"github.com/spf13/cobra"
)

var (
	portKill  bool
	portForce bool
	portAll   bool
)

var portCmd = &cobra.Command{
	Use:   "port <端口号>",
	Short: "检查端口占用，并可按需 kill 占用进程",
	Long: `检查指定 TCP 端口被哪些进程占用，默认仅显示监听(LISTEN)进程，
使用 --all 显示该端口的所有连接。

示例：
  proxyctl port 7892          # 查看 7892 端口占用
  proxyctl port 7892 --kill   # 查看并结束占用进程
  proxyctl port 7892 --force  # 查看并强制结束进程
  proxyctl port 7892 --all    # 显示所有连接（含非监听）`,
	Args: usageArgs(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := portutil.Validate(args[0]); err != nil {
			return usageErr(err)
		}

		procs, err := portutil.Find(args[0], portAll)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if len(procs) == 0 {
			fmt.Fprintf(out, "端口 %s 未被占用。\n", args[0])
			return nil
		}

		fmt.Fprintf(out, "端口 %s 被以下进程占用：\n", args[0])
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PID\tCOMMAND\tADDRESS")
		for _, p := range procs {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.PID, p.Command, p.Address)
		}
		w.Flush()

		if portKill {
			return killAndReport(cmd, procs)
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, "如需结束占用进程，请追加 --kill 参数（强制结束请用 --force）。")
		return nil
	},
}

// killAndReport 结束进程并把结果输出到 stderr（诊断信息），有进程结束失败时返回错误。
func killAndReport(cmd *cobra.Command, procs []portutil.Process) error {
	summary := portutil.Kill(procs, portForce)
	errOut := cmd.ErrOrStderr()
	for _, p := range summary.Killed {
		fmt.Fprintf(errOut, "  ✓ 已结束进程 %s (%s)\n", p.PID, p.Command)
	}
	for _, f := range summary.Failed {
		fmt.Fprintf(errOut, "  ✗ 结束进程 %s (%s) 失败: %v\n", f.Process.PID, f.Process.Command, f.Err)
	}
	if len(summary.Failed) > 0 {
		return errors.New("部分进程未能结束，请检查权限后重试")
	}
	return nil
}

func init() {
	portCmd.Flags().BoolVar(&portKill, "kill", false, "结束占用该端口的进程")
	portCmd.Flags().BoolVar(&portForce, "force", false, "配合 --kill 使用：强制结束进程")
	portCmd.Flags().BoolVar(&portAll, "all", false, "显示该端口的全部连接（默认仅监听）")
}
