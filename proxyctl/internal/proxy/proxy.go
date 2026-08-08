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
