//go:build darwin

package proxy

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseScutilProxy(t *testing.T) {
	out := `<dictionary> {
    HTTPEnable : 1
    HTTPProxy : 127.0.0.1
    HTTPPort : 7890
    HTTPSEnable : 1
    HTTPSProxy : 127.0.0.1
    HTTPSPort : 7891
    SOCKSEnable : 1
    SOCKSProxy : 127.0.0.1
    SOCKSPort : 7892
    ProxyAutoConfigEnable : 1
    ProxyAutoConfigURLString : http://127.0.0.1:8080/proxy.pac
    ProxyAutoDiscoveryEnable : 1
}`
	info, err := parseScutilProxy(strings.NewReader(out))
	if err != nil {
		t.Fatalf("parseScutilProxy: %v", err)
	}
	want := &Info{
		HTTPEnable:    true,
		HTTPHost:      "127.0.0.1",
		HTTPPort:      "7890",
		HTTPSEnable:   true,
		HTTPSHost:     "127.0.0.1",
		HTTPSPort:     "7891",
		SOCKSEnable:   true,
		SOCKSHost:     "127.0.0.1",
		SOCKSPort:     "7892",
		AutoConfig:    true,
		AutoConfigURL: "http://127.0.0.1:8080/proxy.pac",
		AutoDiscovery: true,
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("parseScutilProxy = %#v, want %#v", info, want)
	}
	if got := info.HTTPProxyURL(); got != "http://127.0.0.1:7890" {
		t.Errorf("HTTPProxyURL = %q", got)
	}
	if got := info.PACURL(); got != "http://127.0.0.1:8080/proxy.pac" {
		t.Errorf("PACURL = %q", got)
	}
}

func TestParseScutilProxyDisabled(t *testing.T) {
	out := `HTTPEnable : 0
HTTPSEnable : 0
SOCKSEnable : 0
ProxyAutoConfigEnable : 0`
	info, err := parseScutilProxy(strings.NewReader(out))
	if err != nil {
		t.Fatalf("parseScutilProxy: %v", err)
	}
	want := &Info{}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("parseScutilProxy = %#v, want %#v", info, want)
	}
}

func TestParseScutilProxyPartial(t *testing.T) {
	out := `HTTPEnable : 1
HTTPProxy : proxy.local
HTTPPort : 8080`
	info, err := parseScutilProxy(strings.NewReader(out))
	if err != nil {
		t.Fatalf("parseScutilProxy: %v", err)
	}
	if !info.HTTPEnable || info.HTTPHost != "proxy.local" || info.HTTPPort != "8080" {
		t.Errorf("HTTP 部分解析错误: %#v", info)
	}
	if info.HTTPSEnable || info.SOCKSEnable || info.AutoConfig {
		t.Errorf("未启用字段应保持 false: %#v", info)
	}
}

func TestParseNetworksetupEndpoint(t *testing.T) {
	out := `Enabled: Yes
Server: 127.0.0.1
Port: 7890
Authenticated Proxy Enabled: 0
`
	got, err := parseNetworksetupEndpoint(out)
	if err != nil {
		t.Fatalf("parseNetworksetupEndpoint: %v", err)
	}
	want := EndpointState{Enabled: true, Host: "127.0.0.1", Port: "7890"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseNetworksetupEndpoint = %#v, want %#v", got, want)
	}
}

func TestParseNetworksetupEndpointDisabled(t *testing.T) {
	got, err := parseNetworksetupEndpoint("Enabled: No\n")
	if err != nil {
		t.Fatalf("parseNetworksetupEndpoint: %v", err)
	}
	if got.Enabled || got.Host != "" || got.Port != "" {
		t.Errorf("parseNetworksetupEndpoint = %#v", got)
	}
}

func TestParseNetworksetupPAC(t *testing.T) {
	out := `Enabled: Yes
URL: http://127.0.0.1:8080/proxy.pac
`
	got, err := parseNetworksetupPAC(out)
	if err != nil {
		t.Fatalf("parseNetworksetupPAC: %v", err)
	}
	want := PACState{Enabled: true, URL: "http://127.0.0.1:8080/proxy.pac"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseNetworksetupPAC = %#v, want %#v", got, want)
	}
}
