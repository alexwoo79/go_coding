// Package toolproxy 管理各开发工具的代理配置文件（npm/cargo/pip/docker）。
package toolproxy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultNoProxy 是写入工具配置的默认 no_proxy 值。
const DefaultNoProxy = "localhost,127.0.0.1,::1"

// Tool 表示一个支持持久化代理配置的开发工具。
type Tool interface {
	// Name 返回工具名（小写）。
	Name() string
	// Path 返回配置文件路径。
	Path() string
	// Read 返回当前代理配置（key → 值；nil 或缺失表示未设置）。
	Read() (map[string]*string, error)
	// Write 写入代理配置：key 有值则设置，缺失或 nil 则移除该 key。
	Write(vals map[string]*string) error
	// ProxyValues 返回应用代理时需要写入的 key → 值。
	ProxyValues(proxyURL, noProxy string) map[string]*string
}

// Supported 返回全部支持的工具。
func Supported() []Tool {
	return []Tool{npmTool{}, pipTool{}, cargoTool{}, dockerTool{}}
}

// Find 按名称查找工具（大小写不敏感）。
func Find(name string) (Tool, error) {
	for _, t := range Supported() {
		if strings.EqualFold(t.Name(), name) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("不支持的工具 %q（支持：npm、pip、cargo、docker）", name)
}

// Select 按名称选择工具；未指定时返回全部。
func Select(names []string) ([]Tool, error) {
	if len(names) == 0 {
		return Supported(), nil
	}
	tools := make([]Tool, 0, len(names))
	for _, n := range names {
		t, err := Find(n)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// ReadAll 读取全部工具的当前代理配置。
func ReadAll(tools []Tool) (map[string]map[string]*string, error) {
	out := make(map[string]map[string]*string, len(tools))
	for _, t := range tools {
		vals, err := t.Read()
		if err != nil {
			return nil, err
		}
		out[t.Name()] = vals
	}
	return out, nil
}

// ApplyTo 把代理 URL 写入指定工具。
func ApplyTo(tools []Tool, proxyURL string) error {
	var errs []error
	for _, t := range tools {
		if err := t.Write(t.ProxyValues(proxyURL, DefaultNoProxy)); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", t.Name(), err))
		}
	}
	return errors.Join(errs...)
}

// Clear 清除全部指定工具的代理配置。
func Clear(tools []Tool) error {
	var errs []error
	for _, t := range tools {
		if _, err := os.Stat(t.Path()); os.IsNotExist(err) {
			continue // 配置文件不存在，无需清除
		}
		if err := t.Write(map[string]*string{}); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", t.Name(), err))
		}
	}
	return errors.Join(errs...)
}

// Restore 按快照恢复指定工具（无快照条目的工具视为清除）。
func Restore(tools []Tool, snap map[string]map[string]*string) error {
	var errs []error
	for _, t := range tools {
		vals := snap[t.Name()]
		if vals == nil {
			vals = map[string]*string{}
		}
		hasValue := false
		for _, v := range vals {
			if v != nil {
				hasValue = true
				break
			}
		}
		if !hasValue {
			if _, err := os.Stat(t.Path()); os.IsNotExist(err) {
				continue // 快照为空且文件不存在，无需写入
			}
		}
		if err := t.Write(vals); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", t.Name(), err))
		}
	}
	return errors.Join(errs...)
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
