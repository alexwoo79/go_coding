package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/diagnostic"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "一键诊断开发环境网络状态",
	Long: `组合 status / test / port 做一次完整诊断：检查系统代理、
git 代理、端口监听与网络连通性，并给出修复建议。
发现问题时退出码为 1。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		rep := diagnostic.RunDoctor(ctx)
		out := cmd.OutOrStdout()

		fmt.Fprintln(out, "ProxyCTL Doctor")
		fmt.Fprintln(out, "────────────────────────")

		var sections []string
		seen := make(map[string]bool)
		for _, c := range rep.Checks {
			if !seen[c.Section] {
				seen[c.Section] = true
				sections = append(sections, c.Section)
			}
		}

		var fails, warns int
		for _, sec := range sections {
			fmt.Fprintln(out)
			fmt.Fprintln(out, sec)
			for _, c := range rep.Checks {
				if c.Section != sec {
					continue
				}
				switch c.Status {
				case diagnostic.StatusFail:
					fails++
				case diagnostic.StatusWarn:
					warns++
				}
				fmt.Fprintf(out, "  %-16s %s %s\n", c.Name, c.Status.Mark(), c.Detail)
			}
		}

		if len(rep.Recommendations) > 0 {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Recommendation")
			for _, r := range rep.Recommendations {
				fmt.Fprintf(out, "  → %s\n", r)
			}
		}

		if fails > 0 {
			return fmt.Errorf("发现 %d 个问题（%d 个警告），请参考上方报告", fails, warns)
		}
		return nil
	},
}
