// Package git 封装 git 全局代理配置（http.proxy / https.proxy）的读写。
package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNotSet 表示配置项未设置（对应 git config --global --get 的退出码 1）。
var ErrNotSet = errors.New("git 全局配置未设置该项")

// Runner 抽象 git 命令执行，便于在测试中注入假实现。
type Runner interface {
	Run(args ...string) ([]byte, error)
}

// ExecRunner 使用真实 git 命令执行。
type ExecRunner struct{}

// Run 执行 git 命令；退出码为 1 时返回 ErrNotSet。
func (ExecRunner) Run(args ...string) ([]byte, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return out, ErrNotSet
		}
		return out, err
	}
	return out, nil
}

// Config 提供 git 全局代理配置的读写能力。
type Config struct {
	runner Runner
}

// New 返回使用真实 git 命令的 Config。
func New() *Config {
	return NewWithRunner(ExecRunner{})
}

// NewWithRunner 返回使用自定义 Runner 的 Config，便于测试。
func NewWithRunner(r Runner) *Config {
	return &Config{runner: r}
}

// Keys 是 proxyctl 管理的全部 git 全局配置项。
var Keys = []string{"http.proxy", "https.proxy"}

// Get 读取某个全局配置项；未设置时返回 ("", nil)。
func (c *Config) Get(key string) (string, error) {
	out, err := c.runner.Run("config", "--global", "--get", key)
	if errors.Is(err, ErrNotSet) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("git config --global --get %s 失败: %w", key, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Set 设置某个全局配置项。
func (c *Config) Set(key, value string) error {
	out, err := c.runner.Run("config", "--global", key, value)
	if err != nil {
		return fmt.Errorf("git config --global %s: %w（%s）", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Unset 删除某个全局配置项；未设置时视为成功（幂等）。
func (c *Config) Unset(key string) error {
	_, err := c.runner.Run("config", "--global", "--unset", key)
	if errors.Is(err, ErrNotSet) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("git config --global --unset %s 失败: %w", key, err)
	}
	return nil
}

// Snapshot 记录当前全部受管配置项的值（未设置项不出现）。
func (c *Config) Snapshot() (Snapshot, error) {
	values := make(map[string]string, len(Keys))
	for _, key := range Keys {
		v, err := c.Get(key)
		if err != nil {
			return Snapshot{}, err
		}
		if v != "" {
			values[key] = v
		}
	}
	return Snapshot{Values: values}, nil
}

// Restore 将配置恢复到快照状态：快照中有值的重新设置，没有的删除。
func (c *Config) Restore(s Snapshot) error {
	for _, key := range Keys {
		if v, ok := s.Values[key]; ok && v != "" {
			if err := c.Set(key, v); err != nil {
				return err
			}
			continue
		}
		if err := c.Unset(key); err != nil {
			return err
		}
	}
	return nil
}

// Snapshot 表示 git 全局代理的配置快照。
type Snapshot struct {
	Values map[string]string `json:"values,omitempty"`
}
