//go:build !windows

package port

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Find 通过 lsof 查找占用指定端口的进程。
func Find(port string, all bool) ([]Process, error) {
	args := []string{"-nP", "-iTCP:" + port, "-F", "pcn"}
	if !all {
		args = append(args, "-sTCP:LISTEN")
	}
	out, err := exec.Command("lsof", args...).Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("未找到 lsof 命令，请先安装（macOS 自带；Linux 可用 apt/yum 安装）")
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) == 0 {
			// lsof 未匹配到任何进程时退出码非 0 且无输出，视为正常
			return nil, nil
		}
		return nil, fmt.Errorf("lsof 执行失败: %w", err)
	}

	var procs []Process
	var cur *Process
	seen := make(map[string]bool) // 按 PID 去重

	flush := func() {
		if cur != nil {
			if !seen[cur.PID] {
				seen[cur.PID] = true
				procs = append(procs, *cur)
			}
			cur = nil
		}
	}

	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "p"):
			flush()
			cur = &Process{PID: strings.TrimPrefix(line, "p")}
		case strings.HasPrefix(line, "c"):
			if cur != nil {
				cur.Command = strings.TrimPrefix(line, "c")
			}
		case strings.HasPrefix(line, "n"):
			if cur != nil {
				cur.Address = strings.TrimPrefix(line, "n")
			}
		}
	}
	flush()

	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	return procs, nil
}

// Kill 结束占用端口的所有进程，并返回成功与失败汇总（由调用方负责输出）。
func Kill(procs []Process, force bool) KillSummary {
	var summary KillSummary
	for _, p := range procs {
		args := []string{p.PID}
		if force {
			args = []string{"-9", p.PID}
		}
		if err := exec.Command("kill", args...).Run(); err != nil {
			summary.Failed = append(summary.Failed, KillFailure{Process: p, Err: err})
			continue
		}
		summary.Killed = append(summary.Killed, p)
	}
	return summary
}
