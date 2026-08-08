package cmd

import (
	"github.com/spf13/cobra"
)

// UsageError 表示命令行参数或标志的用法错误。
// main 依据该类型返回退出码 2（Unix 用法错误约定）。
type UsageError struct {
	Err error
}

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

// usageArgs 包装 Cobra 的位置参数校验，使校验错误映射为用法错误（退出码 2）。
func usageArgs(fn cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			return &UsageError{Err: err}
		}
		return nil
	}
}

// usageErr 将普通错误标记为用法错误（退出码 2）。
func usageErr(err error) error {
	return &UsageError{Err: err}
}
