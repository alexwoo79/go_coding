//go:build darwin

package proxy

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// Get 通过 `scutil --proxy` 读取 macOS 系统代理配置。
func Get() (*Info, error) {
	out, err := exec.Command("scutil", "--proxy").Output()
	if err != nil {
		return nil, err
	}

	info := &Info{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
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
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return info, nil
}

// Clear 关闭 macOS 系统代理：遍历所有网络服务，关闭 HTTP/HTTPS/SOCKS 与 PAC 代理。
func Clear() error {
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return fmt.Errorf("获取网络服务列表失败: %w", err)
	}

	var firstErr error
	for _, line := range strings.Split(string(out), "\n") {
		svc := strings.TrimSpace(line)
		// 跳过空行与注释行（首行为 "An asterisk (*) denotes ..."）
		if svc == "" || strings.HasPrefix(svc, "An asterisk") {
			continue
		}
		// 服务名前缀 * 表示已禁用，仍尝试关闭（无害）
		svc = strings.TrimPrefix(svc, "*")

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
