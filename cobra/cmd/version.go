package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// 版本信息，可通过 ldflags 在构建时注入：
//
//	go build -ldflags "\
//	  -X github.com/alexwoo79/go_coding/cobra/cmd.version=1.0.0 \
//	  -X github.com/alexwoo79/go_coding/cobra/cmd.commit=$(git rev-parse --short HEAD) \
//	  -X github.com/alexwoo79/go_coding/cobra/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	  -o proxyctl .
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Args:  usageArgs(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "proxyctl %s (commit: %s, built: %s)\n", version, commit, date)
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Fprintf(out, "go: %s\n", info.GoVersion)
		}
	},
}
