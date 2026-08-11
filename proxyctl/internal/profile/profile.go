// Package profile 管理系统代理 profile 的存取。
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

// EnvDir 可用环境变量覆盖 profile 目录（主要用于测试）。
const EnvDir = "PROXYCTL_PROFILE_DIR"

// Profile 是一组可复用的系统代理端点配置。
type Profile struct {
	Name  string               `json:"name"`
	HTTP  *proxy.EndpointState `json:"http,omitempty"`
	HTTPS *proxy.EndpointState `json:"https,omitempty"`
	SOCKS *proxy.EndpointState `json:"socks,omitempty"`
	PAC   *proxy.PACState      `json:"pac,omitempty"`
}

// Dir 返回 profile 目录（~/.config/proxyctl/profiles，可用 PROXYCTL_PROFILE_DIR 覆盖）。
func Dir() (string, error) {
	if d := os.Getenv(EnvDir); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".config", "proxyctl", "profiles"), nil
}

// Path 返回指定 profile 的文件路径；名称不合法时返回错误。
func Path(name string) (string, error) {
	if !ValidName(name) {
		return "", fmt.Errorf("无效的 profile 名称 %q", name)
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

// ValidName 校验 profile 名称：仅允许字母、数字、-、_、.，且不能为空。
func ValidName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}

// Save 原子写入 profile（先写临时文件再改名），权限为 0600。
func Save(p Profile) error {
	path, err := Path(p.Name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load 读取指定 profile；不存在时返回 (nil, nil)。
func Load(name string) (*Profile, error) {
	path, err := Path(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("解析 profile %q 失败: %w", name, err)
	}
	p.Name = name // 以文件名为准
	return &p, nil
}

// List 返回全部已保存的 profile，按名称排序。
func List() ([]Profile, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Profile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		p, err := Load(name)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Remove 删除指定 profile；不存在时返回错误。
func Remove(name string) error {
	path, err := Path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("未找到 profile %q", name)
		}
		return err
	}
	return nil
}
