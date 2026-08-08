package proxy

import "testing"

func TestProxyURLs(t *testing.T) {
	info := &Info{
		HTTPEnable:  true,
		HTTPHost:    "127.0.0.1",
		HTTPPort:    "7892",
		HTTPSEnable: true,
		HTTPSHost:   "127.0.0.1",
		HTTPSPort:   "7893",
		SOCKSEnable: true,
		SOCKSHost:   "127.0.0.1",
		SOCKSPort:   "7894",
	}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"http", info.HTTPProxyURL(), "http://127.0.0.1:7892"},
		{"https", info.HTTPSProxyURL(), "http://127.0.0.1:7893"},
		{"socks", info.SOCKSPProxyURL(), "socks5h://127.0.0.1:7894"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	disabled := &Info{}
	for name, got := range map[string]string{
		"http":  disabled.HTTPProxyURL(),
		"https": disabled.HTTPSProxyURL(),
		"socks": disabled.SOCKSPProxyURL(),
	} {
		if got != "" {
			t.Errorf("disabled %s URL = %q, want empty", name, got)
		}
	}

	noPort := &Info{HTTPEnable: true, HTTPHost: "proxy.local"}
	if got := noPort.HTTPProxyURL(); got != "http://proxy.local" {
		t.Errorf("no-port URL = %q, want %q", got, "http://proxy.local")
	}
}
