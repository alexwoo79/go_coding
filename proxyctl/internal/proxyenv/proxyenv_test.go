package proxyenv

import (
	"strings"
	"testing"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

func TestFromInfoHTTP(t *testing.T) {
	info := &proxy.Info{
		HTTPEnable: true,
		HTTPHost:   "127.0.0.1",
		HTTPPort:   "7890",
	}
	v := FromInfo(info)
	if v.HTTPProxy == nil || *v.HTTPProxy != "http://127.0.0.1:7890" {
		t.Errorf("HTTPProxy = %v", v.HTTPProxy)
	}
	if v.HTTPSProxy == nil || *v.HTTPSProxy != "http://127.0.0.1:7890" {
		t.Errorf("HTTPSProxy = %v", v.HTTPSProxy)
	}
	if v.AllProxy == nil || *v.AllProxy != "http://127.0.0.1:7890" {
		t.Errorf("AllProxy = %v", v.AllProxy)
	}
	if v.NoProxy == nil || *v.NoProxy != "localhost,127.0.0.1,::1" {
		t.Errorf("NoProxy = %v", v.NoProxy)
	}
}

func TestFromInfoSOCKSFallback(t *testing.T) {
	info := &proxy.Info{
		SOCKSEnable: true,
		SOCKSHost:   "127.0.0.1",
		SOCKSPort:   "7891",
	}
	v := FromInfo(info)
	if v.HTTPProxy == nil || *v.HTTPProxy != "socks5h://127.0.0.1:7891" {
		t.Errorf("HTTPProxy(SOCKS 回退) = %v", v.HTTPProxy)
	}
}

func TestFromInfoDisabled(t *testing.T) {
	v := FromInfo(&proxy.Info{})
	if v.HTTPProxy != nil || v.HTTPSProxy != nil || v.AllProxy != nil || v.NoProxy != nil {
		t.Errorf("禁用时应全部为 nil: %#v", v)
	}
}

func TestScriptPOSIXSet(t *testing.T) {
	url := "http://127.0.0.1:7890"
	noProxy := "localhost,127.0.0.1,::1"
	v := Vars{HTTPProxy: &url, HTTPSProxy: &url, AllProxy: &url, NoProxy: &noProxy}
	s := v.Script(ShellPOSIX)
	for _, want := range []string{
		`export http_proxy="http://127.0.0.1:7890"`,
		`export HTTP_PROXY="http://127.0.0.1:7890"`,
		`export no_proxy="localhost,127.0.0.1,::1"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("脚本缺少 %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "unset") {
		t.Errorf("全部设置时不应有 unset:\n%s", s)
	}
}

func TestScriptPOSIXDirect(t *testing.T) {
	s := Direct().Script(ShellPOSIX)
	for _, name := range varNames {
		want := "unset " + name
		if !strings.Contains(s, want) {
			t.Errorf("脚本缺少 %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "export") {
		t.Errorf("直连模式不应有 export:\n%s", s)
	}
}

func TestScriptFish(t *testing.T) {
	url := "http://127.0.0.1:7890"
	v := Vars{HTTPProxy: &url}
	s := v.Script(ShellFish)
	if !strings.Contains(s, `set -gx http_proxy "http://127.0.0.1:7890"`) {
		t.Errorf("fish 脚本缺少 set 行:\n%s", s)
	}
	if !strings.Contains(s, "set -e https_proxy; or true") {
		t.Errorf("fish 脚本缺少 unset 行:\n%s", s)
	}
}

func TestScriptPowerShell(t *testing.T) {
	url := "http://127.0.0.1:7890"
	v := Vars{HTTPProxy: &url}
	s := v.Script(ShellPowerShell)
	if !strings.Contains(s, `$env:http_proxy = "http://127.0.0.1:7890"`) {
		t.Errorf("powershell 脚本缺少 set 行:\n%s", s)
	}
	if !strings.Contains(s, "Remove-Item Env:https_proxy -ErrorAction SilentlyContinue") {
		t.Errorf("powershell 脚本缺少 Remove-Item 行:\n%s", s)
	}
}
