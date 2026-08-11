package cmd

import (
	"encoding/json"
	"testing"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

func TestBuildStatusJSON(t *testing.T) {
	info := &proxy.Info{
		HTTPEnable:    true,
		HTTPHost:      "127.0.0.1",
		HTTPPort:      "7890",
		AutoConfig:    true,
		AutoConfigURL: "http://127.0.0.1:8080/proxy.pac",
		AutoDiscovery: true,
	}
	out := buildStatusJSON(info, "http://127.0.0.1:7890", "")
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	sys, ok := got["system_proxy"].(map[string]any)
	if !ok {
		t.Fatalf("system_proxy 缺失: %s", data)
	}
	http, ok := sys["http"].(map[string]any)
	if !ok || http["enabled"] != true || http["host"] != "127.0.0.1" || http["port"] != "7890" {
		t.Errorf("system_proxy.http = %#v", sys["http"])
	}
	if _, ok := sys["https"]; ok {
		t.Errorf("未启用的 https 不应出现: %s", data)
	}
	pac, ok := sys["pac"].(map[string]any)
	if !ok || pac["url"] != "http://127.0.0.1:8080/proxy.pac" {
		t.Errorf("system_proxy.pac = %#v", sys["pac"])
	}
	wpad, ok := sys["wpad"].(map[string]any)
	if !ok || wpad["enabled"] != true {
		t.Errorf("system_proxy.wpad = %#v", sys["wpad"])
	}

	git, ok := got["git"].(map[string]any)
	if !ok {
		t.Fatalf("git 缺失: %s", data)
	}
	if git["http_proxy"] != "http://127.0.0.1:7890" {
		t.Errorf("git.http_proxy = %#v", git["http_proxy"])
	}
	if git["https_proxy"] != nil {
		t.Errorf("git.https_proxy = %#v, want null", git["https_proxy"])
	}
}

func TestBuildStatusJSONDisabled(t *testing.T) {
	out := buildStatusJSON(&proxy.Info{}, "", "")
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	sys := got["system_proxy"].(map[string]any)
	for _, k := range []string{"http", "https", "socks", "pac"} {
		if _, ok := sys[k]; ok {
			t.Errorf("未启用时 %s 不应出现: %s", k, data)
		}
	}
	git := got["git"].(map[string]any)
	if git["http_proxy"] != nil || git["https_proxy"] != nil {
		t.Errorf("git 字段应为 null: %s", data)
	}
}
