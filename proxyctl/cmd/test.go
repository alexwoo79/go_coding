package cmd

import (
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
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "执行网络连通性测试",
	Long: `依次执行：公网 IP 检查、百度响应检查、ping 测试、git 连通性测试。
HTTP 请求优先使用系统检测到的代理，未检测到时回退到环境变量或直连。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		client := newTestClient()

		fmt.Fprintln(out, "=== 网络测试 ===")

		// 1. 公网 IP
		fmt.Fprintln(out)
		fmt.Fprintln(out, "1. 检查公网 IP (api.ipify.org)：")
		if ip, err := httpGetBody(client, "https://api.ipify.org"); err != nil {
			fmt.Fprintf(out, "   ✗ 失败: %v\n", err)
		} else {
			fmt.Fprintf(out, "   ✓ 公网 IP: %s\n", ip)
		}

		// 2. 百度响应头
		fmt.Fprintln(out)
		fmt.Fprintln(out, "2. 检查百度响应 (https://www.baidu.com)：")
		resp, err := client.Head("https://www.baidu.com")
		if err != nil {
			fmt.Fprintf(out, "   ✗ 失败: %v\n", err)
		} else {
			fmt.Fprintf(out, "   ✓ 状态码: %s\n", resp.Status)
			resp.Body.Close()
		}

		// 3. ping 测试
		fmt.Fprintln(out)
		fmt.Fprintln(out, "3. ping 测试 (www.baidu.com)：")
		runCmd(out, exec.Command("ping", append(pingCountArgs(4), "www.baidu.com")...))

		// 4. git 测试
		fmt.Fprintln(out)
		fmt.Fprintln(out, "4. git 测试 (github.com)：")
		fmt.Fprintf(out, "   http.proxy  = %s\n", displayValue(gitGet("http.proxy")))
		fmt.Fprintf(out, "   https.proxy = %s\n", displayValue(gitGet("https.proxy")))
		runCmd(out, exec.Command("git", "ls-remote", "https://github.com/github/gitignore.git", "HEAD"))

		return nil
	},
}

// newTestClient 构建一个优先使用系统代理的 HTTP 客户端。
func newTestClient() *http.Client {
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

// httpGetBody 发起 GET 请求并返回响应正文（去除首尾空白）。
func httpGetBody(client *http.Client, rawURL string) (string, error) {
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// runCmd 运行命令并把原始输出写到指定写入器。
func runCmd(w io.Writer, c *exec.Cmd) {
	raw, err := c.CombinedOutput()
	if len(raw) > 0 {
		fmt.Fprint(w, string(raw))
	}
	if err != nil {
		fmt.Fprintf(w, "   ✗ 命令执行失败: %v\n", err)
	}
}

// pingCountArgs 返回 ping 的计数参数（Windows 用 -n，类 Unix 用 -c）。
func pingCountArgs(count int) []string {
	if runtime.GOOS == "windows" {
		return []string{"-n", strconv.Itoa(count)}
	}
	return []string{"-c", strconv.Itoa(count)}
}
