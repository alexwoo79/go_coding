package toolproxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectAndFind(t *testing.T) {
	if len(Supported()) != 4 {
		t.Fatalf("Supported 数量 = %d, want 4", len(Supported()))
	}
	if _, err := Find("npm"); err != nil {
		t.Errorf("Find(npm): %v", err)
	}
	if _, err := Find("NPM"); err != nil {
		t.Errorf("Find(NPM): %v", err)
	}
	if _, err := Find("brew"); err == nil {
		t.Error("Find(brew) 应报错")
	}
	tools, err := Select([]string{"npm", "docker"})
	if err != nil || len(tools) != 2 {
		t.Errorf("Select = %v, %v", tools, err)
	}
}

func TestNPMReadWriteClear(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".npmrc")
	t.Setenv("npm_config_userconfig", rc)
	if err := os.WriteFile(rc, []byte("registry=https://registry.npmmirror.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := npmTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(tool.ProxyValues(url, DefaultNoProxy)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	if !strings.Contains(text, "proxy=http://127.0.0.1:7890") || !strings.Contains(text, "https-proxy=http://127.0.0.1:7890") {
		t.Errorf("npmrc 缺少 proxy:\n%s", text)
	}
	if !strings.Contains(text, "registry=https://registry.npmmirror.com") {
		t.Errorf("npmrc 原有内容丢失:\n%s", text)
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["proxy"] == nil || *vals["proxy"] != url {
		t.Errorf("Read proxy = %v", vals["proxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	text = string(data)
	if strings.Contains(text, "proxy=") {
		t.Errorf("Clear 后仍有 proxy:\n%s", text)
	}
	if !strings.Contains(text, "registry=") {
		t.Errorf("Clear 后原有内容丢失:\n%s", text)
	}
}

func TestPIPSectionReadWrite(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, "pip.conf")
	t.Setenv("PIP_CONFIG_FILE", rc)
	if err := os.WriteFile(rc, []byte("[global]\nindex-url = https://mirrors.aliyun.com/pypi/simple/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := pipTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(map[string]*string{"global.proxy": &url}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	if !strings.Contains(text, "[global]") || !strings.Contains(text, "proxy=http://127.0.0.1:7890") {
		t.Errorf("pip.conf 缺少 [global] proxy:\n%s", text)
	}
	if !strings.Contains(text, "index-url") {
		t.Errorf("pip.conf 原有内容丢失:\n%s", text)
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["global.proxy"] == nil || *vals["global.proxy"] != url {
		t.Errorf("Read global.proxy = %v", vals["global.proxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	if strings.Contains(string(data), "proxy=") {
		t.Errorf("Clear 后仍有 proxy:\n%s", data)
	}
}

func TestCargoTOMLReadWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CARGO_HOME", dir)
	rc := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(rc, []byte("[net]\nretry = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := cargoTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(map[string]*string{"http.proxy": &url}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	if !strings.Contains(text, "[http]") || !strings.Contains(text, `proxy = "http://127.0.0.1:7890"`) {
		t.Errorf("config.toml 缺少 [http] proxy:\n%s", text)
	}
	if !strings.Contains(text, "retry = 2") {
		t.Errorf("config.toml 原有内容丢失:\n%s", text)
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["http.proxy"] == nil || *vals["http.proxy"] != url {
		t.Errorf("Read http.proxy = %v", vals["http.proxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	if strings.Contains(string(data), "proxy") {
		t.Errorf("Clear 后仍有 proxy:\n%s", data)
	}
}
