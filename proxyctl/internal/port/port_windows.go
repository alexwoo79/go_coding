//go:build windows

package port

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Find 通过 netstat 查找占用指定端口的进程，并用 tasklist 回填进程名。
func Find(port string, all bool) ([]Process, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("未找到 netstat 命令")
		}
		return nil, fmt.Errorf("netstat 执行失败: %w", err)
	}

	var procs []Process
	seen := make(map[string]bool) // 按 PID 去重
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Proto") {
			continue
		}
		fields := strings.Fields(line)
		// netstat -ano 列：Proto Local Foreign State PID（TCP 行为 5 列）
		if len(fields) < 5 {
			continue
		}
		local := fields[1]
		state := fields[3]
		pid := fields[4]

		if !strings.HasSuffix(local, ":"+port) {
			continue
		}
		if !all && !strings.EqualFold(state, "LISTENING") {
			continue
		}
		if seen[pid] {
			continue
		}
		seen[pid] = true
		procs = append(procs, Process{
			PID:     pid,
			Command: processName(pid),
			Address: local,
		})
	}

	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	return procs, nil
}

// processName 通过 PowerShell 查询 PID 对应的进程名。
// 进程名以 Base64（UTF-8）编码输出，避免中文等非 UTF-8 区域下
// tasklist 输出的 GBK 或 PowerShell 管道的 UTF-16 编码导致的乱码。
func processName(pid string) string {
	script := fmt.Sprintf(
		"[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes((Get-Process -Id %s -ErrorAction Stop).Name))",
		pid,
	)
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return "?"
	}
	name, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil || len(name) == 0 {
		return "?"
	}
	return string(name)
}

// Kill 通过 taskkill 结束占用端口的所有进程，并返回成功与失败汇总（由调用方负责输出）。
func Kill(procs []Process, force bool) KillSummary {
	var summary KillSummary
	for _, p := range procs {
		args := []string{"/PID", p.PID}
		if force {
			args = append(args, "/F")
		}
		if err := exec.Command("taskkill", args...).Run(); err != nil {
			summary.Failed = append(summary.Failed, KillFailure{Process: p, Err: err})
			continue
		}
		summary.Killed = append(summary.Killed, p)
	}
	return summary
}
