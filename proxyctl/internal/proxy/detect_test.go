package proxy

import "testing"

func TestMatchesProxyKeyword(t *testing.T) {
	for _, ok := range []string{"clash", "Clash", "mihomo", "verge-mihomo", "clash-meta", "v2ray", "xray", "sing-box", "shadowsocks"} {
		if !matchesProxyKeyword(ok) {
			t.Errorf("matchesProxyKeyword(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"node", "PacketTunnel", "nginx", "", "python"} {
		if matchesProxyKeyword(bad) {
			t.Errorf("matchesProxyKeyword(%q) = true, want false", bad)
		}
	}
}

func TestIsSOCKSOnly(t *testing.T) {
	for _, ok := range []string{"ss-local", "shadowsocks-libev", "ssr-local"} {
		if !isSOCKSOnly(ok) {
			t.Errorf("isSOCKSOnly(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"clash", "mihomo", "v2ray"} {
		if isSOCKSOnly(bad) {
			t.Errorf("isSOCKSOnly(%q) = true, want false", bad)
		}
	}
}

func TestDetectedURL(t *testing.T) {
	d := Detected{Scheme: "http", Host: "127.0.0.1", Port: "7890"}
	if d.URL() != "http://127.0.0.1:7890" {
		t.Errorf("URL = %q", d.URL())
	}
	d = Detected{Scheme: "socks5h", Host: "127.0.0.1", Port: "1080"}
	if d.URL() != "socks5h://127.0.0.1:1080" {
		t.Errorf("URL = %q", d.URL())
	}
}
