// Package proxy 封装了读取系统代理配置的逻辑。
//
// 在 macOS 上通过 `scutil --proxy` 读取；其他平台不支持。
package proxy

import "fmt"

// Info 表示系统代理配置信息。
type Info struct {
	HTTPEnable  bool
	HTTPHost    string
	HTTPPort    string
	HTTPSEnable bool
	HTTPSHost   string
	HTTPSPort   string
	SOCKSEnable bool
	SOCKSHost   string
	SOCKSPort   string
	AutoConfig  bool
	// AutoConfigURL 是 PAC 自动代理脚本地址（ProxyAutoConfigURLString）。
	AutoConfigURL string
}

// HTTPProxyURL 返回 HTTP 代理 URL，未启用时返回空字符串。
func (i *Info) HTTPProxyURL() string {
	if !i.HTTPEnable {
		return ""
	}
	return i.proxyURL("http", i.HTTPHost, i.HTTPPort)
}

// HTTPSProxyURL 返回 HTTPS 代理 URL，未启用时返回空字符串。
func (i *Info) HTTPSProxyURL() string {
	if !i.HTTPSEnable {
		return ""
	}
	return i.proxyURL("http", i.HTTPSHost, i.HTTPSPort)
}

// SOCKSPProxyURL 返回 SOCKS5 代理 URL，未启用时返回空字符串。
func (i *Info) SOCKSPProxyURL() string {
	if !i.SOCKSEnable {
		return ""
	}
	return i.proxyURL("socks5h", i.SOCKSHost, i.SOCKSPort)
}

// PACURL 返回 PAC 自动代理脚本地址，未启用时返回空字符串。
func (i *Info) PACURL() string {
	if !i.AutoConfig {
		return ""
	}
	return i.AutoConfigURL
}

// proxyURL 拼接代理 URL；主机为空时返回空字符串，端口为空时省略端口。
func (i *Info) proxyURL(scheme, host, port string) string {
	if host == "" {
		return ""
	}
	if port == "" {
		return scheme + "://" + host
	}
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

// SystemSnapshot 是 clear/restore 需要保留的系统代理状态。
// 各平台填充不同字段：macOS 用 Services，Windows 用 ProxyEnable/ProxyServer/AutoConfigURL。
type SystemSnapshot struct {
	Services      []ServiceState `json:"services,omitempty"`
	ProxyEnable   bool           `json:"proxy_enable"`
	ProxyServer   string         `json:"proxy_server,omitempty"`
	AutoConfigURL string         `json:"auto_config_url,omitempty"`
}

// ServiceState 记录单个网络服务的代理状态（macOS）。
type ServiceState struct {
	Name  string         `json:"name"`
	HTTP  *EndpointState `json:"http,omitempty"`
	HTTPS *EndpointState `json:"https,omitempty"`
	SOCKS *EndpointState `json:"socks,omitempty"`
	PAC   *PACState      `json:"pac,omitempty"`
}

// EndpointState 记录单个 HTTP/HTTPS/SOCKS 代理端点的状态。
type EndpointState struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host,omitempty"`
	Port    string `json:"port,omitempty"`
}

// PACState 记录 PAC 自动代理的状态与脚本地址。
type PACState struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}
