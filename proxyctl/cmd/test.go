package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/diagnostic"
	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
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
		ctx := context.Background()
		client := diagnostic.NewHTTPClient()

		gitCfg := git.New()
		httpVal, httpErr := gitCfg.Get("http.proxy")
		httpsVal, httpsErr := gitCfg.Get("https.proxy")

		checks := []diagnostic.Checker{
			diagnostic.PublicIPCheck{Client: client, URL: "https://api.ipify.org"},
			diagnostic.HTTPCheck{Client: client, URL: "https://www.baidu.com"},
			diagnostic.PingCheck{Target: "www.baidu.com", Count: 4},
			diagnostic.GitCheck{Target: "https://github.com/github/gitignore.git"},
		}

		fmt.Fprintln(out, "=== 网络测试 ===")
		for i, res := range diagnostic.RunAll(ctx, checks) {
			fmt.Fprintln(out)
			fmt.Fprintf(out, "%d. %s\n", i+1, res.Name)
			if strings.HasPrefix(res.Name, "git ") {
				fmt.Fprintf(out, "   http.proxy  = %s\n", displayValueOrErr(httpVal, httpErr))
				fmt.Fprintf(out, "   https.proxy = %s\n", displayValueOrErr(httpsVal, httpsErr))
			}
			elapsed := res.Duration.Round(time.Millisecond)
			if res.OK {
				fmt.Fprintf(out, "   ✓ %s (用时 %s)\n", res.Detail, elapsed)
			} else {
				detail := res.Detail
				if res.Err != nil {
					detail = res.Err.Error()
				}
				fmt.Fprintf(out, "   ✗ 失败: %s (用时 %s)\n", detail, elapsed)
			}
		}
		return nil
	},
}

// displayValueOrErr 将 git 配置读取结果转换为可显示的文本。
func displayValueOrErr(v string, err error) string {
	if err != nil {
		return "读取失败: " + err.Error()
	}
	return displayValue(v)
}
