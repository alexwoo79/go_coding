// Package port 封装了检查端口占用与结束占用进程的逻辑。
// Unix 平台使用 lsof/kill，Windows 平台使用 netstat/tasklist/taskkill。
package port

import (
	"fmt"
	"strconv"
)

// Process 表示占用某端口的进程；一个进程可能同时占用多个地址。
type Process struct {
	PID       int
	Command   string
	Addresses []string
}

// KillFailure 记录单个进程结束失败的原因。
type KillFailure struct {
	Process Process
	Err     error
}

// KillSummary 汇总 Kill 的结果，由调用方负责输出。
type KillSummary struct {
	Killed []Process
	Failed []KillFailure
}

// Validate 校验端口号是否合法。
func Validate(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("无效端口号 %q：请输入 1-65535 之间的数字", port)
	}
	return nil
}

// contains 判断字符串切片是否包含指定元素。
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
