package toolproxy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// StateVersion 是工具快照文件格式版本。
	StateVersion = 1
	// EnvStateFile 可用环境变量覆盖工具快照路径（主要用于测试）。
	EnvStateFile = "PROXYCTL_TOOLS_STATE_FILE"
)

// Snapshot 是工具代理配置的快照（key → 值；nil 表示未设置）。
type Snapshot struct {
	Version   int                           `json:"version"`
	Timestamp time.Time                     `json:"timestamp"`
	Tools     map[string]map[string]*string `json:"tools,omitempty"`
}

// DefaultStatePath 返回默认工具快照路径（~/.config/proxyctl/tools-state.json）。
func DefaultStatePath() (string, error) {
	if p := os.Getenv(EnvStateFile); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".config", "proxyctl", "tools-state.json"), nil
}

// SaveState 原子写入工具快照。
func SaveState(path string, s *Snapshot) error {
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

// LoadState 读取工具快照；文件不存在时返回 (nil, nil)。
func LoadState(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析工具状态文件 %s 失败: %w", path, err)
	}
	return &s, nil
}
