package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// executeCommand 以指定参数运行根命令并捕获输出。
func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRootHelp(t *testing.T) {
	got, err := executeCommand()
	if err != nil {
		t.Fatalf("root help: unexpected error: %v", err)
	}
	if !strings.Contains(got, "proxyctl") {
		t.Errorf("root help 输出缺少命令名: %q", got)
	}
}

func TestVersionCommand(t *testing.T) {
	got, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version: unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, "proxyctl "+version) {
		t.Errorf("version 输出 = %q, 期望前缀 %q", got, "proxyctl "+version)
	}
	if !strings.Contains(got, "go: ") {
		t.Errorf("version 输出缺少 go 版本: %q", got)
	}
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"port 缺少参数", []string{"port"}},
		{"port 参数非数字", []string{"port", "abc"}},
		{"port 参数超出范围", []string{"port", "70000"}},
		{"status 多余参数", []string{"status", "extra"}},
		{"未知子命令", []string{"nosuch"}},
		{"未知标志", []string{"status", "--bogus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(tt.args...)
			if err == nil {
				t.Fatalf("args %v: 期望错误，实际无错误", tt.args)
			}
			var ue *UsageError
			if !errors.As(err, &ue) {
				t.Errorf("args %v: 错误 %v 不是 UsageError", tt.args, err)
			}
		})
	}
}

func TestDisplayValue(t *testing.T) {
	if got := displayValue(""); got != "<未设置>" {
		t.Errorf("displayValue(\"\") = %q", got)
	}
	const url = "http://127.0.0.1:7892"
	if got := displayValue(url); got != url {
		t.Errorf("displayValue(%q) = %q", url, got)
	}
}
