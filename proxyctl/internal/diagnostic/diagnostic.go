// Package diagnostic 提供网络连通性检查与 doctor 诊断编排。
package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

// CheckResult 是单个检查的结果。
type CheckResult struct {
	Name     string
	OK       bool
	Detail   string
	Duration time.Duration
	Err      error
}

// Checker 表示一次可执行的诊断检查。
type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// RunAll 顺序执行全部检查并返回结果。
func RunAll(ctx context.Context, checkers []Checker) []CheckResult {
	results := make([]CheckResult, 0, len(checkers))
	for _, c := range checkers {
		results = append(results, c.Check(ctx))
	}
	return results
}

// NewHTTPClient 构建一个优先使用系统代理的 HTTP 客户端。
// 代理选择顺序：HTTP → HTTPS → SOCKS；均未启用时回退到环境变量或直连。
// 对回环地址（127.0.0.1 / localhost / ::1）与本地服务端口（Ollama 11434、cc-switch 15721 等）
// 始终直连，避免本地服务经代理访问失败。
func NewHTTPClient() *http.Client {
	proxyFunc := http.ProxyFromEnvironment
	if info, err := proxy.Get(); err == nil {
		for _, u := range []string{
			info.HTTPProxyURL(),
			info.HTTPSProxyURL(),
			info.SOCKSPProxyURL(),
		} {
			if u == "" {
				continue
			}
			if pu, perr := url.Parse(u); perr == nil {
				proxyFunc = http.ProxyURL(pu)
				break
			}
		}
	}
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: bypassLocalProxy(proxyFunc),
		},
	}
}

// localBypassPorts 是需要绕过代理的本地服务端口：Ollama / llama-server 等 LLM 服务，
// 以及 cc-switch（15721）。
var localBypassPorts = map[string]struct{}{
	"11434": {},
	"11435": {},
	"8080":  {},
	"15721": {},
}

// bypassLocalProxy 包装 proxyFunc：目标是回环地址或本地服务端口时返回 nil（直连），其余透传。
func bypassLocalProxy(proxyFunc func(*http.Request) (*url.URL, error)) func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		if shouldBypassProxy(req) {
			return nil, nil
		}
		return proxyFunc(req)
	}
}

// shouldBypassProxy 判断请求是否应绕过代理（回环地址或本地服务端口）。
func shouldBypassProxy(req *http.Request) bool {
	if isLoopbackHost(req.URL.Hostname()) {
		return true
	}
	_, ok := localBypassPorts[req.URL.Port()]
	return ok
}

// isLoopbackHost 判断主机名是否为回环地址（localhost、127.0.0.0/8 或 ::1）。
func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// PublicIPCheck 通过外部服务查询公网 IP。
type PublicIPCheck struct {
	Client *http.Client
	URL    string
}

// Name 返回检查名称。
func (c PublicIPCheck) Name() string { return "公网 IP" }

// Check 执行公网 IP 查询。
func (c PublicIPCheck) Check(ctx context.Context) CheckResult {
	res := CheckResult{Name: c.Name()}
	start := time.Now()
	defer func() { res.Duration = time.Since(start) }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		res.Err = err
		return res
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		res.Err = err
		return res
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		res.Err = err
		return res
	}
	ip := strings.TrimSpace(string(body))
	if ip == "" {
		res.Err = errors.New("响应为空")
		return res
	}
	res.OK = true
	res.Detail = ip
	return res
}

// HTTPCheck 通过 HEAD 请求检查目标站点的可访问性。
type HTTPCheck struct {
	Client *http.Client
	URL    string
}

// Name 返回检查名称。
func (c HTTPCheck) Name() string { return "HTTP " + c.URL }

// Check 执行 HEAD 请求。
func (c HTTPCheck) Check(ctx context.Context) CheckResult {
	res := CheckResult{Name: c.Name()}
	start := time.Now()
	defer func() { res.Duration = time.Since(start) }()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.URL, nil)
	if err != nil {
		res.Err = err
		return res
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		res.Err = err
		return res
	}
	resp.Body.Close()
	res.OK = true
	res.Detail = resp.Status
	return res
}

// PingCheck 通过系统 ping 命令检查主机连通性。
type PingCheck struct {
	Target string
	Count  int
}

// Name 返回检查名称。
func (c PingCheck) Name() string { return "ping " + c.Target }

// Check 执行 ping 命令。
func (c PingCheck) Check(ctx context.Context) CheckResult {
	res := CheckResult{Name: c.Name()}
	start := time.Now()
	defer func() { res.Duration = time.Since(start) }()

	args := pingCountArgs(c.Count)
	args = append(args, c.Target)
	cmd := exec.CommandContext(ctx, "ping", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.Err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		return res
	}
	res.OK = true
	res.Detail = firstNonEmptyLine(string(out))
	return res
}

// GitCheck 通过 git ls-remote 检查远端仓库连通性。
type GitCheck struct {
	Target string
}

// Name 返回检查名称。
func (c GitCheck) Name() string { return "git " + c.Target }

// Check 执行 git ls-remote。
func (c GitCheck) Check(ctx context.Context) CheckResult {
	res := CheckResult{Name: c.Name()}
	start := time.Now()
	defer func() { res.Duration = time.Since(start) }()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", c.Target, "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.Err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		return res
	}
	res.OK = true
	res.Detail = firstNonEmptyLine(string(out))
	return res
}

// pingCountArgs 返回 ping 的计数参数（Windows 用 -n，类 Unix 用 -c）。
func pingCountArgs(count int) []string {
	if runtime.GOOS == "windows" {
		return []string{"-n", strconv.Itoa(count)}
	}
	return []string{"-c", strconv.Itoa(count)}
}

// firstNonEmptyLine 返回输出的第一行非空内容。
func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
