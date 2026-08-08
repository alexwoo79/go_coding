package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// gitGet 读取某个 git 全局配置项，未设置时返回空字符串。
func gitGet(key string) string {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitSet 设置 git 全局配置项。
func gitSet(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config --global %s: %w（%s）", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gitUnset 删除 git 全局配置项。
func gitUnset(key string) error {
	return exec.Command("git", "config", "--global", "--unset", key).Run()
}

// displayValue 将空字符串显示为“<未设置>”。
func displayValue(v string) string {
	if v == "" {
		return "<未设置>"
	}
	return v
}
