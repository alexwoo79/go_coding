// Package diagnostic 提供网络连通性检查与 doctor 诊断编排。
package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"io"
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
func NewHTTPClient() *http.Client {
	transport := &http.Transport{}
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
				transport.Proxy = http.ProxyURL(pu)
				break
			}
		}
	}
	if transport.Proxy == nil {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}
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
