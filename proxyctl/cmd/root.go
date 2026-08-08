package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "proxyctl",
	Short: "macOS 代理与端口管理 CLI 工具",
	Long: `proxyctl 是一个基于 macOS 系统代理设置的命令行工具，
可以检测系统代理、自动配置 git 代理、执行网络连通性测试，
以及检查端口占用并结束占用进程。

用法示例：
  proxyctl status    # 查看当前系统代理与 git 代理状态
  proxyctl apply     # 根据系统代理自动设置 git 全局代理
  proxyctl clear     # 关闭系统代理并清除 git 代理
  proxyctl test      # 执行网络连通性测试
  proxyctl port 7892 # 检查端口占用（--kill 可结束进程）`,
	Version:       version,
	Args:          usageArgs(cobra.NoArgs),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute 执行根命令并返回错误，由 main 负责退出处理。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 标志解析错误属于用法错误，由 main 映射为退出码 2。
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &UsageError{Err: err}
	})

	rootCmd.AddCommand(
		statusCmd,
		applyCmd,
		clearCmd,
		testCmd,
		portCmd,
		versionCmd,
	)
}
