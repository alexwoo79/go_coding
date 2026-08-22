package toolproxy

import (
	"os"
	"path/filepath"
	"strings"
)

// brewProxyKeys 是 brew.env 中由 proxyctl 管理的代理变量（Homebrew 支持标准小写形式）。
var brewProxyKeys = []string{"http_proxy", "https_proxy", "all_proxy", "no_proxy"}

// brewTool 管理用户级 Homebrew 环境文件 brew.env 中的代理变量。
//
// Homebrew 的 bin/brew 启动时会读取该文件，并导出其中形如 http_proxy= 的行
// （匹配 all_proxy / no_proxy / ftp_proxy / http_proxy / https_proxy）。
// 位置与 bin/brew 一致：
//   - $HOMEBREW_XDG_CONFIG_HOME/homebrew/brew.env
//   - $XDG_CONFIG_HOME/homebrew/brew.env（设置了 XDG_CONFIG_HOME 时）
//   - 默认 ~/.homebrew/brew.env
type brewTool struct{}

func (brewTool) Name() string { return "brew" }

func (brewTool) Path() string {
	if p := os.Getenv("HOMEBREW_XDG_CONFIG_HOME"); p != "" {
		return filepath.Join(p, "homebrew", "brew.env")
	}
	if p := os.Getenv("XDG_CONFIG_HOME"); p != "" {
		return filepath.Join(p, "homebrew", "brew.env")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".homebrew", "brew.env")
}

func (t brewTool) Read() (map[string]*string, error) {
	out := map[string]*string{}
	text, err := readFileOrEmpty(t.Path())
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(text, "\n") {
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || !containsKey(brewProxyKeys, key) {
			continue
		}
		v := strings.TrimSpace(value)
		out[key] = &v
	}
	return out, nil
}

func (t brewTool) Write(vals map[string]*string) error {
	text, err := readFileOrEmpty(t.Path())
	if err != nil {
		return err
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		key, _, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if ok && containsKey(brewProxyKeys, key) {
			if v, has := vals[key]; has && v != nil {
				out = append(out, key+"="+*v)
			}
			continue
		}
		out = append(out, line)
	}
	for _, k := range brewProxyKeys {
		if v, has := vals[k]; has && v != nil {
			out = append(out, k+"="+*v)
		}
	}
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	return atomicWrite(t.Path(), []byte(strings.Join(out, "\n")+"\n"))
}

func (brewTool) ProxyValues(proxyURL, noProxy string) map[string]*string {
	u := &proxyURL
	return map[string]*string{
		"http_proxy":  u,
		"https_proxy": u,
		"all_proxy":   u,
		"no_proxy":    &noProxy,
	}
}
