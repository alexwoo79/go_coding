package diagnostic

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/port"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

// DoctorStatus 表示单条诊断检查的状态。
type DoctorStatus int

const (
	// StatusOK 表示正常。
	StatusOK DoctorStatus = iota
	// StatusWarn 表示需要留意但不阻塞。
	StatusWarn
	// StatusFail 表示存在问题。
	StatusFail
	// StatusSkip 表示当前平台不支持、跳过。
	StatusSkip
)

// Mark 返回状态符号。
func (s DoctorStatus) Mark() string {
	switch s {
	case StatusOK:
		return "✓"
	case StatusWarn:
		return "!"
	case StatusFail:
		return "✗"
	default:
		return "-"
	}
}

// DoctorCheck 是单条诊断结果。
type DoctorCheck struct {
	Section string
	Name    string
	Status  DoctorStatus
	Detail  string
}

// DoctorReport 是 doctor 的完整结果与修复建议。
type DoctorReport struct {
	Checks          []DoctorCheck
	Recommendations []string
}

// RunDoctor 执行一次完整的开发环境网络诊断（只读）。
func RunDoctor(ctx context.Context) *DoctorReport {
	rep := &DoctorReport{}
	add := func(section, name string, status DoctorStatus, detail string) {
		rep.Checks = append(rep.Checks, DoctorCheck{Section: section, Name: name, Status: status, Detail: detail})
	}
	recommend := func(msg string) {
		for _, r := range rep.Recommendations {
			if r == msg {
				return
			}
		}
		rep.Recommendations = append(rep.Recommendations, msg)
	}

	// --- System ---
	systemSupported := runtime.GOOS == "darwin" || runtime.GOOS == "windows"
	add("System", "platform", StatusOK, runtime.GOOS)
	if !systemSupported {
		add("System", "system proxy", StatusSkip, "当前平台不支持系统代理管理")
	}
	for _, tool := range systemTools() {
		path, err := exec.LookPath(tool)
		if err != nil {
			add("System", tool, StatusFail, "未找到命令")
			continue
		}
		add("System", tool, StatusOK, path)
	}

	// --- Proxy ---
	var info *proxy.Info
	if !systemSupported {
		add("Proxy", "系统代理", StatusSkip, "当前平台不支持")
	} else {
		var err error
		info, err = proxy.Get()
		if err != nil {
			add("Proxy", "读取系统代理", StatusFail, err.Error())
		} else {
			addEndpointCheck := func(name string, enabled bool, u string) {
				switch {
				case !enabled:
					add("Proxy", name, StatusOK, "未启用")
				case u == "":
					add("Proxy", name, StatusWarn, "已启用但地址缺失")
				default:
					add("Proxy", name, StatusOK, u)
				}
			}
			addEndpointCheck("HTTP", info.HTTPEnable, info.HTTPProxyURL())
			addEndpointCheck("HTTPS", info.HTTPSEnable, info.HTTPSProxyURL())
			addEndpointCheck("SOCKS", info.SOCKSEnable, info.SOCKSPProxyURL())
			addEndpointCheck("PAC", info.AutoConfig, info.PACURL())
		}
	}

	// --- Git ---
	gitCfg := git.New()
	httpProxy, httpErr := gitCfg.Get("http.proxy")
	httpsProxy, httpsErr := gitCfg.Get("https.proxy")
	addGit := func(key, value string, err error) {
		switch {
		case err != nil:
			add("Git", key, StatusFail, "读取失败: "+err.Error())
		case value == "":
			add("Git", key, StatusOK, "未设置")
		default:
			add("Git", key, StatusOK, value)
		}
	}
	addGit("http.proxy", httpProxy, httpErr)
	addGit("https.proxy", httpsProxy, httpsErr)
	if httpErr == nil && httpsErr == nil {
		if httpProxy != "" && httpsProxy == "" {
			recommend("git http.proxy 已设置但 https.proxy 未设置：运行 proxyctl apply 补齐")
		}
		if httpProxy == "" && httpsProxy != "" {
			recommend("git https.proxy 已设置但 http.proxy 未设置：运行 proxyctl apply 补齐")
		}
	}

	// --- Ports ---
	ports := make(map[string][]string)
	addPort := func(p, src string) {
		if p == "" {
			return
		}
		ports[p] = append(ports[p], src)
	}
	if httpProxy != "" {
		addPort(proxyPort(httpProxy), "git http.proxy")
	}
	if httpsProxy != "" {
		addPort(proxyPort(httpsProxy), "git https.proxy")
	}
	if info != nil {
		if info.HTTPEnable {
			addPort(info.HTTPPort, "系统 HTTP")
		}
		if info.HTTPSEnable {
			addPort(info.HTTPSPort, "系统 HTTPS")
		}
		if info.SOCKSEnable {
			addPort(info.SOCKSPort, "系统 SOCKS")
		}
	}
	if len(ports) == 0 {
		add("Ports", "代理端口", StatusSkip, "未检测到需要检查的代理端口")
	} else {
		sortedPorts := make([]string, 0, len(ports))
		for p := range ports {
			sortedPorts = append(sortedPorts, p)
		}
		sort.Slice(sortedPorts, func(i, j int) bool {
			return atoi(sortedPorts[i]) < atoi(sortedPorts[j])
		})
		for _, p := range sortedPorts {
			procs, err := port.Find(p, false)
			if err != nil {
				add("Ports", p, StatusFail, "检查失败: "+err.Error())
				continue
			}
			if len(procs) == 0 {
				src := strings.Join(ports[p], "、")
				add("Ports", p, StatusFail, "未监听（被 "+src+" 引用）")
				recommend(fmt.Sprintf("端口 %s 未监听，但 %s 指向它：请确认代理程序已启动，或运行 proxyctl clear", p, src))
				continue
			}
			names := make([]string, 0, len(procs))
			for _, pr := range procs {
				names = append(names, pr.Command)
			}
			add("Ports", p, StatusOK, "LISTEN ("+strings.Join(names, ", ")+")")
		}
	}

	// --- 一致性 ---
	systemEnabled := info != nil && (info.HTTPEnable || info.HTTPSEnable || info.SOCKSEnable || info.AutoConfig)
	if systemEnabled && httpProxy == "" && httpsProxy == "" && httpErr == nil && httpsErr == nil {
		recommend("系统代理已启用但 git 代理未配置：运行 proxyctl apply")
	}
	if !systemEnabled && (httpProxy != "" || httpsProxy != "") {
		val := httpProxy
		if val == "" {
			val = httpsProxy
		}
		recommend("系统代理未启用但 git 代理仍指向 " + val + "：运行 proxyctl clear 清除")
	}

	// --- Connectivity ---
	if _, err := net.DefaultResolver.LookupHost(ctx, "github.com"); err != nil {
		add("Connectivity", "DNS", StatusFail, err.Error())
	} else {
		add("Connectivity", "DNS", StatusOK, "github.com 解析成功")
	}

	client := NewHTTPClient()
	httpRes := HTTPCheck{Client: client, URL: "https://www.baidu.com"}.Check(ctx)
	if httpRes.OK {
		add("Connectivity", "Internet", StatusOK, httpRes.Detail)
	} else {
		add("Connectivity", "Internet", StatusFail, httpRes.Err.Error())
	}

	gitRes := GitCheck{Target: "https://github.com/github/gitignore.git"}.Check(ctx)
	if gitRes.OK {
		add("Connectivity", "GitHub", StatusOK, gitRes.Detail)
	} else {
		add("Connectivity", "GitHub", StatusFail, gitRes.Err.Error())
	}

	// --- Environment ---
	for _, k := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY"} {
		if v := os.Getenv(k); v != "" {
			add("Environment", k, StatusWarn, v)
			recommend(fmt.Sprintf("环境变量 %s 已设置（%s），会影响 curl 等工具；proxyctl 不管理这些变量", k, v))
		} else {
			add("Environment", k, StatusOK, "未设置")
		}
	}

	return rep
}

// systemTools 返回按平台需要检查的外部命令。
func systemTools() []string {
	tools := []string{"git", "ping"}
	switch runtime.GOOS {
	case "darwin":
		tools = append(tools, "networksetup", "scutil", "lsof")
	case "windows":
		tools = append(tools, "netstat", "tasklist")
	}
	return tools
}

// proxyPort 从代理 URL 中提取端口。
func proxyPort(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Port()
}

// atoi 忽略错误地解析整数，用于端口排序。
func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
