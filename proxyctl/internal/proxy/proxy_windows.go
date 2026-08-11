//go:build windows

package proxy

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

// Get 从 Windows 注册表读取系统代理配置。
func Get() (*Info, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	info := &Info{}

	if enable, _, err := key.GetIntegerValue("ProxyEnable"); err == nil {
		enabled := enable == 1
		info.HTTPEnable = enabled
		info.HTTPSEnable = enabled
		info.SOCKSEnable = enabled
	}

	if url, _, err := key.GetStringValue("AutoConfigURL"); err == nil && url != "" {
		info.AutoConfig = true
		info.AutoConfigURL = url
	}

	server, _, err := key.GetStringValue("ProxyServer")
	if err != nil || server == "" {
		return info, nil
	}
	parseProxyServer(info, server)
	return info, nil
}

// parseProxyServer 解析 Windows 的 ProxyServer 注册表值。
// 支持两种格式：
//   - "host:port"（所有协议共用）
//   - "http=host:port;https=host:port;socks=host:port"（分协议）
func parseProxyServer(info *Info, server string) {
	server = strings.TrimSpace(server)
	if server == "" {
		return
	}
	if !strings.Contains(server, "=") {
		host, port := splitHostPort(server)
		info.HTTPHost, info.HTTPPort = host, port
		info.HTTPSHost, info.HTTPSPort = host, port
		return
	}
	for _, part := range strings.Split(server, ";") {
		part = strings.TrimSpace(part)
		proto, addr, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		host, port := splitHostPort(addr)
		switch strings.ToLower(strings.TrimSpace(proto)) {
		case "http":
			info.HTTPHost, info.HTTPPort = host, port
		case "https":
			info.HTTPSHost, info.HTTPSPort = host, port
		case "socks", "socks5":
			info.SOCKSHost, info.SOCKSPort = host, port
		}
	}
}

// splitHostPort 将 "host:port" 拆分为主机与端口。
func splitHostPort(addr string) (host, port string) {
	addr = strings.TrimSpace(addr)
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[:i], addr[i+1:]
	}
	return addr, ""
}

// Clear 关闭 Windows 系统代理（将 ProxyEnable 置为 0）。
func Clear() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetDWordValue("ProxyEnable", 0)
}

// SnapshotSystem 记录 Windows 注册表中的系统代理状态。
func SnapshotSystem() (SystemSnapshot, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.QUERY_VALUE)
	if err != nil {
		return SystemSnapshot{}, err
	}
	defer key.Close()

	snap := SystemSnapshot{}
	if enable, _, err := key.GetIntegerValue("ProxyEnable"); err == nil {
		snap.ProxyEnable = enable == 1
	}
	snap.ProxyServer, _, _ = key.GetStringValue("ProxyServer")
	snap.AutoConfigURL, _, _ = key.GetStringValue("AutoConfigURL")
	return snap, nil
}

// RestoreSystem 恢复 Windows 注册表中的系统代理状态。
func RestoreSystem(snap SystemSnapshot) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	v := uint32(0)
	if snap.ProxyEnable {
		v = 1
	}
	if err := key.SetDWordValue("ProxyEnable", v); err != nil {
		return err
	}
	if snap.ProxyServer != "" {
		if err := key.SetStringValue("ProxyServer", snap.ProxyServer); err != nil {
			return err
		}
	}
	if snap.AutoConfigURL != "" {
		if err := key.SetStringValue("AutoConfigURL", snap.AutoConfigURL); err != nil {
			return err
		}
	}
	return nil
}
