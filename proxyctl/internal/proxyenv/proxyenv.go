// Package proxyenv 生成把系统代理映射为终端环境变量的脚本。
package proxyenv

import (
	"fmt"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

// Shell 表示目标 shell 语法。
type Shell string

const (
	// ShellPOSIX 覆盖 zsh / bash / sh / dash。
	ShellPOSIX Shell = "posix"
	// ShellFish 是 fish shell。
	ShellFish Shell = "fish"
	// ShellPowerShell 是 Windows PowerShell / pwsh。
	ShellPowerShell Shell = "powershell"
)

// varNames 是脚本管理的全部环境变量名（小写 + 大写，覆盖两类工具）。
var varNames = []string{
	"http_proxy", "HTTP_PROXY",
	"https_proxy", "HTTPS_PROXY",
	"all_proxy", "ALL_PROXY",
	"no_proxy", "NO_PROXY",
}

// Vars 是一组代理环境变量；nil 表示未设置（脚本中执行 unset）。
type Vars struct {
	HTTPProxy  *string
	HTTPSProxy *string
	AllProxy   *string
	NoProxy    *string
}

// FromInfo 从系统代理信息生成环境变量。
// 与 apply 一致：优先 HTTP 代理，回退 SOCKS；未检测到代理时返回全部未设置（直连）。
func FromInfo(info *proxy.Info) Vars {
	u := info.HTTPProxyURL()
	if u == "" {
		u = info.SOCKSPProxyURL()
	}
	if u == "" {
		return Vars{}
	}
	return Vars{
		HTTPProxy:  &u,
		HTTPSProxy: &u,
		AllProxy:   &u,
		NoProxy:    strPtr("localhost,127.0.0.1,::1"),
	}
}

// Direct 返回直连模式（全部未设置）。
func Direct() Vars {
	return Vars{}
}

func strPtr(s string) *string { return &s }

// Script 生成指定 shell 的配置脚本。
func (v Vars) Script(shell Shell) string {
	switch shell {
	case ShellFish:
		return v.fish()
	case ShellPowerShell:
		return v.powershell()
	default:
		return v.posix()
	}
}

// values 返回变量名到值的映射（nil 表示 unset）。
func (v Vars) values() map[string]*string {
	return map[string]*string{
		"http_proxy":  v.HTTPProxy,
		"HTTP_PROXY":  v.HTTPProxy,
		"https_proxy": v.HTTPSProxy,
		"HTTPS_PROXY": v.HTTPSProxy,
		"all_proxy":   v.AllProxy,
		"ALL_PROXY":   v.AllProxy,
		"no_proxy":    v.NoProxy,
		"NO_PROXY":    v.NoProxy,
	}
}

func (v Vars) posix() string {
	var b strings.Builder
	fmt.Fprintln(&b, "# proxyctl env")
	vals := v.values()
	for _, name := range varNames {
		if val := vals[name]; val != nil {
			fmt.Fprintf(&b, "export %s=\"%s\"\n", name, *val)
		} else {
			fmt.Fprintf(&b, "unset %s 2>/dev/null || true\n", name)
		}
	}
	return b.String()
}

func (v Vars) fish() string {
	var b strings.Builder
	fmt.Fprintln(&b, "# proxyctl env")
	vals := v.values()
	for _, name := range varNames {
		if val := vals[name]; val != nil {
			fmt.Fprintf(&b, "set -gx %s \"%s\"\n", name, *val)
		} else {
			fmt.Fprintf(&b, "set -e %s; or true\n", name)
		}
	}
	return b.String()
}

func (v Vars) powershell() string {
	var b strings.Builder
	fmt.Fprintln(&b, "# proxyctl env")
	vals := v.values()
	for _, name := range varNames {
		if val := vals[name]; val != nil {
			fmt.Fprintf(&b, "$env:%s = \"%s\"\n", name, *val)
		} else {
			fmt.Fprintf(&b, "Remove-Item Env:%s -ErrorAction SilentlyContinue\n", name)
		}
	}
	return b.String()
}
