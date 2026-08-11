package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvInstallRemove(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	orig := "alias ll='ls -l'\n"
	if err := os.WriteFile(rc, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := envInstallCmd.Flags().Set("file", rc); err != nil {
		t.Fatal(err)
	}
	got, err := executeCommand("env", "install")
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !strings.Contains(got, "已写入") {
		t.Errorf("install 输出 = %q", got)
	}
	content, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, envHookStart) || !strings.Contains(text, "proxyctl env --shell") {
		t.Errorf("hook 未写入:\n%s", text)
	}
	if !strings.Contains(text, orig) {
		t.Errorf("原内容丢失:\n%s", text)
	}
	if _, err := os.Stat(rc + ".proxyctl.bak"); err != nil {
		t.Errorf("备份文件缺失: %v", err)
	}

	// 幂等：重复安装不产生重复块
	got, err = executeCommand("env", "install")
	if err != nil {
		t.Fatalf("重复 install: %v", err)
	}
	if !strings.Contains(got, "无需重复安装") {
		t.Errorf("重复 install 输出 = %q", got)
	}
	content, _ = os.ReadFile(rc)
	if n := strings.Count(string(content), envHookStart); n != 1 {
		t.Errorf("hook 出现 %d 次，应为 1 次", n)
	}

	// remove
	got, err = executeCommand("env", "remove")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(got, "已从") {
		t.Errorf("remove 输出 = %q", got)
	}
	content, _ = os.ReadFile(rc)
	text = string(content)
	if strings.Contains(text, envHookStart) {
		t.Errorf("remove 后仍有 hook:\n%s", text)
	}
	if !strings.Contains(text, orig) {
		t.Errorf("remove 后原内容丢失:\n%s", text)
	}
}

func TestEnvRemoveNotInstalled(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(rc, []byte("export FOO=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := envRemoveCmd.Flags().Set("file", rc); err != nil {
		t.Fatal(err)
	}
	_, err := executeCommand("env", "remove")
	if err == nil {
		t.Fatal("未安装时 remove 应报错")
	}
}

func TestEnvInstallNewFile(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".zshrc")
	if err := envInstallCmd.Flags().Set("file", rc); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCommand("env", "install"); err != nil {
		t.Fatalf("install: %v", err)
	}
	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("新文件未创建: %v", err)
	}
	if !strings.Contains(string(data), envHookStart) {
		t.Errorf("新文件缺少 hook:\n%s", data)
	}
}

func TestRemoveManagedBlock(t *testing.T) {
	orig := "# comment\n" + envHookStart + "\nfoo\n" + envHookEnd + "\ntail\n"
	got, ok := removeManagedBlock(orig)
	if !ok {
		t.Fatal("应找到受管块")
	}
	if got != "# comment\ntail\n" {
		t.Errorf("removeManagedBlock = %q", got)
	}
}
