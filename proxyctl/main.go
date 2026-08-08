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
