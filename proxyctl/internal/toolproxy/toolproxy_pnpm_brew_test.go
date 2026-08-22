package toolproxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPNPMYAMLReadWriteClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	rc := filepath.Join(dir, "pnpm", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rc, []byte("saveExact: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := pnpmTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(tool.ProxyValues(url, DefaultNoProxy)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	for _, want := range []string{
		"httpProxy: http://127.0.0.1:7890",
		"httpsProxy: http://127.0.0.1:7890",
		"noProxy: " + DefaultNoProxy,
		"saveExact: true",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("config.yaml 缺少 %q:\n%s", want, text)
		}
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["httpProxy"] == nil || *vals["httpProxy"] != url {
		t.Errorf("Read httpProxy = %v", vals["httpProxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	text = string(data)
	for _, key := range pnpmKeys {
		if strings.Contains(text, key+":") {
			t.Errorf("Clear 后仍有 %s:\n%s", key, text)
		}
	}
	if !strings.Contains(text, "saveExact: true") {
		t.Errorf("Clear 后原有内容丢失:\n%s", text)
	}
}

func TestPNPMYAMLQuotedValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	rc := filepath.Join(dir, "pnpm", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rc, []byte(`httpProxy: "http://old:8080"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := pnpmTool{}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["httpProxy"] == nil || *vals["httpProxy"] != "http://old:8080" {
		t.Errorf("Read 引号值 = %v", vals["httpProxy"])
	}
}

func TestBrewEnvReadWriteClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOMEBREW_XDG_CONFIG_HOME", dir)
	rc := filepath.Join(dir, "homebrew", "brew.env")
	if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rc, []byte("HOMEBREW_NO_AUTO_UPDATE=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := brewTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(tool.ProxyValues(url, DefaultNoProxy)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	for _, want := range []string{
		"http_proxy=" + url,
		"https_proxy=" + url,
		"all_proxy=" + url,
		"no_proxy=" + DefaultNoProxy,
		"HOMEBREW_NO_AUTO_UPDATE=1",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("brew.env 缺少 %q:\n%s", want, text)
		}
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["https_proxy"] == nil || *vals["https_proxy"] != url {
		t.Errorf("Read https_proxy = %v", vals["https_proxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	text = string(data)
	for _, key := range brewProxyKeys {
		if strings.Contains(text, key+"=") {
			t.Errorf("Clear 后仍有 %s:\n%s", key, text)
		}
	}
	if !strings.Contains(text, "HOMEBREW_NO_AUTO_UPDATE=1") {
		t.Errorf("Clear 后原有内容丢失:\n%s", text)
	}
}
