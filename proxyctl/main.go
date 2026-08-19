// Package proxyctl 提供了一个命令行工具 proxyctl，用于在 macOS、Windows 和 Linux 上管理系统代理和 git 代理配置。
// 它支持自动检测系统代理并设置 git 全局代理，清除系统和 git 代理，以及保存和恢复代理状态快照。
// 该工具使用 cobra 库实现命令行接口，提供了 apply、clear、restore 等子命令，并处理用法错误和运行时错误。
// 主要功能包括：
// 1. 自动设置 git 全局代理：根据当前系统代理设置 git 的 http.proxy 和 https.proxy。
// 2. 清除系统和 git 代理：关闭系统代理并清除 git 全局代理配置。
// 3. 保存和恢复代理状态快照：在执行 apply 或 clear 前保存当前状态，之后可通过 restore 恢复。
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alexwoo79/go_coding/proxyctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		// 用法错误（无效参数/标志）返回退出码 2，其余运行时错误返回退出码 1。
		var ue *cmd.UsageError
		if errors.As(err, &ue) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
