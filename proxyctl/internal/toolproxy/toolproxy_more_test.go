package toolproxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerJSONReadWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	rc := filepath.Join(dir, "config.json")
	orig := "{\n  \"auths\": {\n    \"https://index.docker.io/v1/\": {}\n  }\n}\n"
	if err := os.WriteFile(rc, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := dockerTool{}
	url := "http://127.0.0.1:7890"
	if err := tool.Write(tool.ProxyValues(url, DefaultNoProxy)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(rc)
	text := string(data)
	for _, want := range []string{
		`"httpProxy": "http://127.0.0.1:7890"`,
		`"httpsProxy": "http://127.0.0.1:7890"`,
		`"noProxy": "localhost,127.0.0.1,::1"`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("config.json 缺少 %s:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "auths") {
		t.Errorf("config.json 原有 auths 丢失:\n%s", text)
	}
	vals, err := tool.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if vals["proxies.default.httpProxy"] == nil || *vals["proxies.default.httpProxy"] != url {
		t.Errorf("Read httpProxy = %v", vals["proxies.default.httpProxy"])
	}
	if err := tool.Write(map[string]*string{}); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	data, _ = os.ReadFile(rc)
	text = string(data)
	if strings.Contains(text, "proxies") {
		t.Errorf("Clear 后仍有 proxies:\n%s", text)
	}
	if !strings.Contains(text, "auths") {
		t.Errorf("Clear 后 auths 丢失:\n%s", text)
	}
}

func TestApplyClearRestore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("npm_config_userconfig", filepath.Join(dir, ".npmrc"))
	t.Setenv("PIP_CONFIG_FILE", filepath.Join(dir, "pip.conf"))
	t.Setenv("CARGO_HOME", dir)
	t.Setenv("DOCKER_CONFIG", dir)

	tools, err := Select(nil)
	if err != nil {
		t.Fatal(err)
	}
	url := "http://127.0.0.1:7890"
	if err := ApplyTo(tools, url); err != nil {
		t.Fatalf("ApplyTo: %v", err)
	}
	all, err := ReadAll(tools)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	for name, vals := range all {
		if len(vals) == 0 {
			t.Errorf("工具 %s 未写入代理", name)
		}
	}
	snap := map[string]map[string]*string{}
	for name, vals := range all {
		snap[name] = vals
	}
	if err := Clear(tools); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	all, _ = ReadAll(tools)
	for name, vals := range all {
		if len(vals) != 0 {
			t.Errorf("工具 %s clear 后仍有配置: %v", name, vals)
		}
	}
	if err := Restore(tools, snap); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	all, _ = ReadAll(tools)
	for name, vals := range all {
		if len(vals) == 0 {
			t.Errorf("工具 %s 未恢复", name)
		}
	}
}

func TestStateSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tools-state.json")
	t.Setenv(EnvStateFile, path)
	url := "http://127.0.0.1:7890"
	want := &Snapshot{
		Version: StateVersion,
		Tools: map[string]map[string]*string{
			"npm": {"proxy": &url},
		},
	}
	if err := SaveState(path, want); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if got == nil || got.Tools["npm"]["proxy"] == nil || *got.Tools["npm"]["proxy"] != url {
		t.Errorf("LoadState = %#v", got)
	}
}
