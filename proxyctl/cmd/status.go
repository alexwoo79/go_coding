package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/spf13/cobra"
)

var statusJSON bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前 macOS 系统代理与 git 代理状态",
	Args:  usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := proxy.Get()
		if err != nil {
			return err
		}

		gitCfg := git.New()
		httpVal, httpErr := gitCfg.Get("http.proxy")
		httpsVal, httpsErr := gitCfg.Get("https.proxy")

		out := cmd.OutOrStdout()
		if statusJSON {
			if httpErr != nil {
				return httpErr
			}
			if httpsErr != nil {
				return httpsErr
			}
			data, err := json.MarshalIndent(buildStatusJSON(info, httpVal, httpsVal), "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(out, string(data))
			return nil
		}

		fmt.Fprintln(out, "=== macOS 系统代理 ===")
		fmt.Fprintln(out, proxyLine("HTTP", info.HTTPEnable, info.HTTPProxyURL()))
		fmt.Fprintln(out, proxyLine("HTTPS", info.HTTPSEnable, info.HTTPSProxyURL()))
		fmt.Fprintln(out, proxyLine("SOCKS", info.SOCKSEnable, info.SOCKSPProxyURL()))
		fmt.Fprintln(out, proxyLine("PAC", info.AutoConfig, info.PACURL()))

		fmt.Fprintln(out)
		fmt.Fprintln(out, "=== Git 全局代理 ===")
		if httpErr != nil {
			fmt.Fprintf(out, "  http.proxy  = 读取失败: %v\n", httpErr)
		} else {
			fmt.Fprintf(out, "  http.proxy  = %s\n", displayValue(httpVal))
		}
		if httpsErr != nil {
			fmt.Fprintf(out, "  https.proxy = 读取失败: %v\n", httpsErr)
		} else {
			fmt.Fprintf(out, "  https.proxy = %s\n", displayValue(httpsVal))
		}

		return nil
	},
}

// statusJSONOutput 是 status --json 的输出结构，便于 AI Agent 消费。
type statusJSONOutput struct {
	SystemProxy systemProxyJSON `json:"system_proxy"`
	Git         gitProxyJSON    `json:"git"`
}

type systemProxyJSON struct {
	HTTP  *endpointJSON `json:"http,omitempty"`
	HTTPS *endpointJSON `json:"https,omitempty"`
	SOCKS *endpointJSON `json:"socks,omitempty"`
	PAC   *pacJSON      `json:"pac,omitempty"`
}

type endpointJSON struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host,omitempty"`
	Port    string `json:"port,omitempty"`
}

type pacJSON struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

type gitProxyJSON struct {
	HTTPProxy  *string `json:"http_proxy"`
	HTTPSProxy *string `json:"https_proxy"`
}

// buildStatusJSON 把系统代理与 git 代理状态转换为 JSON 输出结构。
func buildStatusJSON(info *proxy.Info, httpVal, httpsVal string) statusJSONOutput {
	out := statusJSONOutput{}
	if info != nil {
		if info.HTTPEnable {
			out.SystemProxy.HTTP = &endpointJSON{Enabled: true, Host: info.HTTPHost, Port: info.HTTPPort}
		}
		if info.HTTPSEnable {
			out.SystemProxy.HTTPS = &endpointJSON{Enabled: true, Host: info.HTTPSHost, Port: info.HTTPSPort}
		}
		if info.SOCKSEnable {
			out.SystemProxy.SOCKS = &endpointJSON{Enabled: true, Host: info.SOCKSHost, Port: info.SOCKSPort}
		}
		if info.AutoConfig {
			out.SystemProxy.PAC = &pacJSON{Enabled: true, URL: info.AutoConfigURL}
		}
	}
	if httpVal != "" {
		v := httpVal
		out.Git.HTTPProxy = &v
	}
	if httpsVal != "" {
		v := httpsVal
		out.Git.HTTPSProxy = &v
	}
	return out
}

// proxyLine 生成单行代理状态文本。
func proxyLine(name string, enabled bool, url string) string {
	if !enabled {
		return fmt.Sprintf("%-7s代理: 未启用", name)
	}
	if url != "" {
		return fmt.Sprintf("%-7s代理已启用: %s", name, url)
	}
	return fmt.Sprintf("%-7s代理: 已启用", name)
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "以 JSON 格式输出")
}
