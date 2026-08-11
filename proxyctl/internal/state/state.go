// Package state 负责 proxyctl 状态快照的存取。
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

const (
	// Version 是快照文件格式版本。
	Version = 1
	// EnvFile 可用环境变量覆盖快照文件路径（主要用于测试）。
	EnvFile = "PROXYCTL_STATE_FILE"
)

// State 是 proxyctl apply 之前保存的系统代理与 git 代理状态。
type State struct {
	Version   int                  `json:"version"`
	Timestamp time.Time            `json:"timestamp"`
	System    proxy.SystemSnapshot `json:"system_proxy"`
	Git       git.Snapshot         `json:"git"`
}

// DefaultPath 返回默认快照文件路径（~/.config/proxyctl/state.json），
// 可通过 PROXYCTL_STATE_FILE 环境变量覆盖。
func DefaultPath() (string, error) {
	if p := os.Getenv(EnvFile); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".config", "proxyctl", "state.json"), nil
}

// Save 原子写入快照（先写临时文件再改名），权限为 0600。
func Save(path string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
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

// Load 读取快照；文件不存在时返回 (nil, nil)。
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析状态文件 %s 失败: %w", path, err)
	}
	return &s, nil
}

// Remove 删除快照文件；文件不存在时视为成功。
func Remove(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
