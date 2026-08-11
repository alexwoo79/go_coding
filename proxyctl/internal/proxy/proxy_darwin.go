//go:build darwin

package proxy

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Get 通过 `scutil --proxy` 读取 macOS 系统代理配置。
func Get() (*Info, error) {
	out, err := exec.Command("scutil", "--proxy").Output()
	if err != nil {
		return nil, err
	}
	return parseScutilProxy(strings.NewReader(string(out)))
}

// parseScutilProxy 解析 `scutil --proxy` 的输出。
func parseScutilProxy(r io.Reader) (*Info, error) {
	info := &Info{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 输出形如 "  HTTPPort : 7892"
		key, value, ok := strings.Cut(line, " : ")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "HTTPEnable":
			info.HTTPEnable = value == "1"
		case "HTTPProxy":
			info.HTTPHost = value
		case "HTTPPort":
			info.HTTPPort = value
		case "HTTPSEnable":
			info.HTTPSEnable = value == "1"
		case "HTTPSProxy":
			info.HTTPSHost = value
		case "HTTPSPort":
			info.HTTPSPort = value
		case "SOCKSEnable":
			info.SOCKSEnable = value == "1"
		case "SOCKSProxy":
			info.SOCKSHost = value
		case "SOCKSPort":
			info.SOCKSPort = value
		case "ProxyAutoConfigEnable":
			info.AutoConfig = value == "1"
		case "ProxyAutoConfigURLString":
			info.AutoConfigURL = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return info, nil
}

// Clear 关闭 macOS 系统代理：遍历所有网络服务，关闭 HTTP/HTTPS/SOCKS 与 PAC 代理。
func Clear() error {
	services, err := listServices()
	if err != nil {
		return err
	}

	var firstErr error
	for _, svc := range services {
		for _, stateCmd := range []string{
			"-setwebproxystate",
			"-setsecurewebproxystate",
			"-setsocksfirewallproxystate",
			"-setautoproxystate",
		} {
			if err := exec.Command("networksetup", stateCmd, svc, "off").Run(); err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("关闭服务 %q 的代理失败: %w", svc, err)
				}
			}
		}
	}
	return firstErr
}

// listServices 返回全部网络服务名（跳过注释行，去掉已禁用标记 *）。
func listServices() ([]string, error) {
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil, fmt.Errorf("获取网络服务列表失败: %w", err)
	}
	var services []string
	for _, line := range strings.Split(string(out), "\n") {
		svc := strings.TrimSpace(line)
		// 跳过空行与注释行（首行为 "An asterisk (*) denotes ..."）
		if svc == "" || strings.HasPrefix(svc, "An asterisk") {
			continue
		}
		// 服务名前缀 * 表示已禁用，仍尝试操作（无害）
		services = append(services, strings.TrimPrefix(svc, "*"))
	}
	return services, nil
}

// SnapshotSystem 记录每个网络服务的代理状态，供 restore 恢复。
func SnapshotSystem() (SystemSnapshot, error) {
	services, err := listServices()
	if err != nil {
		return SystemSnapshot{}, err
	}

	snap := SystemSnapshot{}
	for _, svc := range services {
		st := ServiceState{Name: svc}
		if out, err := exec.Command("networksetup", "-getwebproxy", svc).Output(); err == nil {
			if e, perr := parseNetworksetupEndpoint(string(out)); perr == nil {
				st.HTTP = &e
			}
		}
		if out, err := exec.Command("networksetup", "-getsecurewebproxy", svc).Output(); err == nil {
			if e, perr := parseNetworksetupEndpoint(string(out)); perr == nil {
				st.HTTPS = &e
			}
		}
		if out, err := exec.Command("networksetup", "-getsocksfirewallproxy", svc).Output(); err == nil {
			if e, perr := parseNetworksetupEndpoint(string(out)); perr == nil {
				st.SOCKS = &e
			}
		}
		if out, err := exec.Command("networksetup", "-getautoproxy", svc).Output(); err == nil {
			if p, perr := parseNetworksetupPAC(string(out)); perr == nil {
				st.PAC = &p
			}
		}
		snap.Services = append(snap.Services, st)
	}
	return snap, nil
}

// RestoreSystem 按快照恢复每个网络服务的代理状态。
func RestoreSystem(snap SystemSnapshot) error {
	var firstErr error
	for _, st := range snap.Services {
		if st.HTTP != nil {
			restoreEndpoint("web", st.Name, *st.HTTP, &firstErr)
		}
		if st.HTTPS != nil {
			restoreEndpoint("secureweb", st.Name, *st.HTTPS, &firstErr)
		}
		if st.SOCKS != nil {
			restoreEndpoint("socksfirewall", st.Name, *st.SOCKS, &firstErr)
		}
		if st.PAC != nil {
			if st.PAC.Enabled && st.PAC.URL != "" {
				if err := exec.Command("networksetup", "-setautoproxyurl", st.Name, st.PAC.URL).Run(); err != nil && firstErr == nil {
					firstErr = fmt.Errorf("恢复服务 %q 的 PAC 代理失败: %w", st.Name, err)
				}
			} else if err := exec.Command("networksetup", "-setautoproxystate", st.Name, "off").Run(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("关闭服务 %q 的 PAC 代理失败: %w", st.Name, err)
			}
		}
	}
	return firstErr
}

// restoreEndpoint 恢复单个 HTTP/HTTPS/SOCKS 端点：快照启用时设置地址并开启，否则关闭。
func restoreEndpoint(kind, svc string, e EndpointState, firstErr *error) {
	if e.Enabled && e.Host != "" && e.Port != "" {
		args := []string{"-set" + kind + "proxy", svc, e.Host, e.Port}
		if err := exec.Command("networksetup", args...).Run(); err != nil {
			if *firstErr == nil {
				*firstErr = fmt.Errorf("设置服务 %q 的 %s 代理失败: %w", svc, kind, err)
			}
			return
		}
		if err := exec.Command("networksetup", "-set"+kind+"proxystate", svc, "on").Run(); err != nil && *firstErr == nil {
			*firstErr = fmt.Errorf("启用服务 %q 的 %s 代理失败: %w", svc, kind, err)
		}
		return
	}
	if err := exec.Command("networksetup", "-set"+kind+"proxystate", svc, "off").Run(); err != nil && *firstErr == nil {
		*firstErr = fmt.Errorf("关闭服务 %q 的 %s 代理失败: %w", svc, kind, err)
	}
}

// parseNetworksetupEndpoint 解析 `networksetup -getwebproxy / -getsecurewebproxy / -getsocksfirewallproxy` 输出。
func parseNetworksetupEndpoint(out string) (EndpointState, error) {
	var e EndpointState
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "Enabled":
			e.Enabled = value == "Yes"
		case "Server":
			e.Host = value
		case "Port":
			e.Port = value
		}
	}
	return e, scanner.Err()
}

// parseNetworksetupPAC 解析 `networksetup -getautoproxy` 输出。
func parseNetworksetupPAC(out string) (PACState, error) {
	var p PACState
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "Enabled":
			p.Enabled = value == "Yes"
		case "URL":
			p.URL = value
		}
	}
	return p, scanner.Err()
}
