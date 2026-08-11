package proxy

import (
	"fmt"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/port"
)

// proxyKeywords 是常见代理程序的进程名关键字。
var proxyKeywords = []string{
	"clash", "mihomo", "verge", "v2ray", "xray",
	"sing-box", "singbox", "shadowsocks", "ss-local", "ssr",
	"trojan", "hysteria", "naive", "surge", "stash", "piglite",
}

// socksOnlyKeywords 是仅提供 SOCKS 代理的程序关键字。
var socksOnlyKeywords = []string{"ss-local", "ssr", "shadowsocks"}

// probePorts 是常见代理监听端口（按优先级）。
var probePorts = []string{"7890", "7891", "7892", "1080", "1087", "10808", "10809", "20171", "8889", "8118"}

// Detected 表示检测到的代理程序端点。
type Detected struct {
	Scheme  string // http / socks5h
	Host    string
	Port    string
	Command string
	Source  string
}

// URL 返回代理 URL。
func (d Detected) URL() string { return d.Scheme + "://" + d.Host + ":" + d.Port }

// Detect 检测运行中的代理程序及其端口。
// 优先扫描常见端口并匹配代理程序进程名；未找到时回退到系统代理配置。
func Detect() (Detected, error) {
	for _, p := range probePorts {
		procs, err := port.Find(p, false)
		if err != nil || len(procs) == 0 {
			continue
		}
		for _, proc := range procs {
			if matchesProxyKeyword(proc.Command) {
				scheme := "http"
				if isSOCKSOnly(proc.Command) {
					scheme = "socks5h"
				}
				return Detected{
					Scheme:  scheme,
					Host:    "127.0.0.1",
					Port:    p,
					Command: proc.Command,
					Source:  "自动检测",
				}, nil
			}
		}
	}

	// 回退：系统代理已启用时直接使用其配置。
	if info, err := Get(); err == nil {
		if u := info.HTTPProxyURL(); u != "" {
			return Detected{Scheme: "http", Host: info.HTTPHost, Port: info.HTTPPort, Command: "系统代理", Source: "系统代理"}, nil
		}
		if u := info.SOCKSPProxyURL(); u != "" {
			return Detected{Scheme: "socks5h", Host: info.SOCKSHost, Port: info.SOCKSPort, Command: "系统代理", Source: "系统代理"}, nil
		}
	}

	return Detected{}, fmt.Errorf(
		"未检测到运行中的代理程序（已检查端口 %s）；请先启动 Clash/Mihomo 等代理，或使用 proxyctl profile use <名称>",
		strings.Join(probePorts, ", "),
	)
}

// matchesProxyKeyword 判断进程名是否匹配常见代理程序。
func matchesProxyKeyword(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(command))
	for _, k := range proxyKeywords {
		if strings.Contains(cmd, k) {
			return true
		}
	}
	return false
}

// isSOCKSOnly 判断进程名是否仅提供 SOCKS 代理。
func isSOCKSOnly(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(command))
	for _, k := range socksOnlyKeywords {
		if strings.Contains(cmd, k) {
			return true
		}
	}
	return false
}
