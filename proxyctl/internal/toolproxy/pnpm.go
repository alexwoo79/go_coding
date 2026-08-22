package toolproxy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// pnpmKeys 是 pnpm 全局配置中的代理键（camelCase，pnpm v11 config.yaml）。
var pnpmKeys = []string{"httpProxy", "httpsProxy", "noProxy"}

// pnpmTool 管理 pnpm 全局配置 config.yaml 的 httpProxy / httpsProxy / noProxy。
//
// pnpm v11 起配置迁移到 YAML 文件，位置参考官方文档：
//   - $XDG_CONFIG_HOME/pnpm/config.yaml（设置了 XDG_CONFIG_HOME 时）
//   - macOS:   ~/Library/Preferences/pnpm/config.yaml
//   - Linux:   ~/.config/pnpm/config.yaml
//   - Windows: ~/AppData/Local/pnpm/config/config.yaml
type pnpmTool struct{}

func (pnpmTool) Name() string { return "pnpm" }

func (pnpmTool) Path() string {
	if p := os.Getenv("XDG_CONFIG_HOME"); p != "" {
		return filepath.Join(p, "pnpm", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local", "pnpm", "config", "config.yaml")
	case "darwin":
		return filepath.Join(home, "Library", "Preferences", "pnpm", "config.yaml")
	default:
		return filepath.Join(home, ".config", "pnpm", "config.yaml")
	}
}

func (t pnpmTool) Read() (map[string]*string, error) {
	return readYAMLKeyValues(t.Path(), pnpmKeys)
}

func (t pnpmTool) Write(vals map[string]*string) error {
	return writeYAMLKeyValues(t.Path(), pnpmKeys, vals)
}

func (pnpmTool) ProxyValues(proxyURL, noProxy string) map[string]*string {
	u := &proxyURL
	return map[string]*string{
		"httpProxy":  u,
		"httpsProxy": u,
		"noProxy":    &noProxy,
	}
}

// readYAMLKeyValues 读取 YAML 键值对文件中指定键的值；未设置的键不出现。
func readYAMLKeyValues(path string, keys []string) (map[string]*string, error) {
	out := map[string]*string{}
	text, err := readFileOrEmpty(path)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(text, "\n") {
		key, value, ok := splitYAMLKeyValue(line)
		if !ok || !containsKey(keys, key) {
			continue
		}
		v := strings.TrimSpace(unquoteYAML(value))
		out[key] = &v
	}
	return out, nil
}

// writeYAMLKeyValues 写回 YAML 键值对文件：保留其他行，vals 中缺失或为 nil 的键被移除。
func writeYAMLKeyValues(path string, keys []string, vals map[string]*string) error {
	text, err := readFileOrEmpty(path)
	if err != nil {
		return err
	}
	var out []string
	written := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		key, _, ok := splitYAMLKeyValue(line)
		if ok && containsKey(keys, key) {
			if v, has := vals[key]; has && v != nil {
				out = append(out, key+": "+yamlQuote(*v))
				written[key] = true
			}
			continue
		}
		out = append(out, line)
	}
	for _, k := range keys {
		if v, has := vals[k]; has && v != nil && !written[k] {
			out = append(out, k+": "+yamlQuote(*v))
		}
	}
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	return atomicWrite(path, []byte(strings.Join(out, "\n")+"\n"))
}

// splitYAMLKeyValue 解析 `key: value` 形式的 YAML 行；键仅限字母数字（pnpm 配置键是 camelCase）。
func splitYAMLKeyValue(line string) (key, value string, ok bool) {
	trimmed := strings.TrimSpace(line)
	idx := strings.Index(trimmed, ":")
	if idx <= 0 || !isYAMLKey(trimmed[:idx]) {
		return "", "", false
	}
	return trimmed[:idx], strings.TrimSpace(trimmed[idx+1:]), true
}

func isYAMLKey(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// unquoteYAML 去除单双引号包裹的值。
func unquoteYAML(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// yamlQuote 仅在必要时给值加双引号，避免破坏 YAML 解析。
func yamlQuote(v string) string {
	if v == "" || strings.Contains(v, ": ") || strings.Contains(v, " #") {
		return `"` + v + `"`
	}
	return v
}
