//go:build !windows

package port

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
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
	return parseLsof(string(out)), nil
}

// parseLsof 解析 `lsof -F pcn` 输出；同一进程的多个地址合并保留。
func parseLsof(out string) []Process {
	var procs []Process
	byPID := make(map[int]int)
	cur := -1

	for _, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "p"):
			pid, err := strconv.Atoi(strings.TrimPrefix(line, "p"))
			if err != nil {
				cur = -1
				continue
			}
			if i, ok := byPID[pid]; ok {
				cur = i
				continue
			}
			byPID[pid] = len(procs)
			procs = append(procs, Process{PID: pid})
			cur = len(procs) - 1
		case strings.HasPrefix(line, "c"):
			if cur >= 0 && procs[cur].Command == "" {
				procs[cur].Command = strings.TrimPrefix(line, "c")
			}
		case strings.HasPrefix(line, "n"):
			if cur >= 0 {
				addr := strings.TrimPrefix(line, "n")
				if !contains(procs[cur].Addresses, addr) {
					procs[cur].Addresses = append(procs[cur].Addresses, addr)
				}
			}
		}
	}

	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	for i := range procs {
		sort.Strings(procs[i].Addresses)
	}
	return procs
}

// Kill 结束占用端口的所有进程，并返回成功与失败汇总（由调用方负责输出）。
func Kill(procs []Process, force bool) KillSummary {
	var summary KillSummary
	for _, p := range procs {
		args := []string{strconv.Itoa(p.PID)}
		if force {
			args = []string{"-9", strconv.Itoa(p.PID)}
		}
		if err := exec.Command("kill", args...).Run(); err != nil {
			summary.Failed = append(summary.Failed, KillFailure{Process: p, Err: err})
			continue
		}
		summary.Killed = append(summary.Killed, p)
	}
	return summary
}
