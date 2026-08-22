package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxyenv"
	"github.com/spf13/cobra"
)

var (
	envShellFlag string
	envClearFlag bool
	envFileFlag  string
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "生成/安装终端代理环境变量",
	Long: `把当前系统代理生成为终端环境变量脚本（http_proxy/https_proxy/all_proxy/no_proxy
	及大写形式），让 curl、pip、npm、cargo、brew、uv 等 CLI 工具统一走代理或直连。

用法：
  proxyctl env                # 输出脚本，配合 eval 使用：eval "$(proxyctl env)"
  proxyctl env --clear        # 输出直连（unset 全部代理变量）
  proxyctl env install        # 写入 shell 配置文件，新开终端自动生效
  proxyctl env remove         # 移除 install 写入的配置

配合 profile 使用不同网络环境：
  proxyctl profile use clash  # 切换到某个网络环境（新开终端自动跟随）
  proxyctl profile use direct # 切换直连`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell, _, err := resolveShell(envShellFlag)
		if err != nil {
			return usageErr(err)
		}
		info, err := proxy.Get()
		if err != nil {
			return err
		}
		var vars proxyenv.Vars
		if envClearFlag {
			vars = proxyenv.Direct()
		} else {
			vars = proxyenv.FromInfo(info)
		}
		fmt.Fprint(cmd.OutOrStdout(), vars.Script(shell))
		return nil
	},
}

var envInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "把代理环境变量 hook 写入 shell 配置文件",
	Long: `在 shell 配置文件（默认 ~/.zshrc，可用 --file 指定）中写入一段受管 hook，
每次新开终端自动执行 proxyctl env，跟随当前系统代理（含 profile 切换）。
hook 带标记块，重复安装幂等，首次安装前会备份原文件为 <文件>.proxyctl.bak。`,
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell, shellName, err := resolveShell(envShellFlag)
		if err != nil {
			return usageErr(err)
		}
		rc, err := shellRCFile(envFileFlag, shellName)
		if err != nil {
			return err
		}
		installed, err := installEnvHook(rc, shellName, shell)
		if err != nil {
			return fmt.Errorf("写入 %s 失败: %w", rc, err)
		}
		out := cmd.OutOrStdout()
		if !installed {
			fmt.Fprintf(out, "%s 已包含 proxyctl env hook，无需重复安装。\n", rc)
			return nil
		}
		fmt.Fprintf(out, "已写入 %s。新开终端将自动应用当前系统代理；\n", rc)
		fmt.Fprintln(out, "切换网络环境请用 proxyctl profile use <名称>。")
		return nil
	},
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "移除 install 写入的代理环境变量 hook",
	Args:  usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, shellName, err := resolveShell(envShellFlag)
		if err != nil {
			return usageErr(err)
		}
		rc, err := shellRCFile(envFileFlag, shellName)
		if err != nil {
			return err
		}
		removed, err := removeEnvHook(rc)
		if err != nil {
			return fmt.Errorf("移除 %s 中的 hook 失败: %w", rc, err)
		}
		if !removed {
			return fmt.Errorf("%s 中未找到 proxyctl env hook（可能尚未安装）", rc)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已从 %s 移除 proxyctl env hook。\n", rc)
		return nil
	},
}

const (
	envHookStart = "# >>> proxyctl env (managed, do not edit) >>>"
	envHookEnd   = "# <<< proxyctl env <<<"
)

// resolveShell 解析 --shell 参数；未指定时按 $SHELL 推断。
func resolveShell(name string) (proxyenv.Shell, string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = shellBase()
	}
	switch name {
	case "fish":
		return proxyenv.ShellFish, "fish", nil
	case "powershell", "pwsh":
		return proxyenv.ShellPowerShell, "powershell", nil
	case "zsh", "bash", "sh", "dash", "ksh", "posix":
		return proxyenv.ShellPOSIX, name, nil
	}
	return "", "", fmt.Errorf("不支持的 shell %q（支持 zsh/bash/sh/fish/powershell）", name)
}

// shellBase 返回 $SHELL 的 basename，无法识别时默认 zsh。
func shellBase() string {
	if s := os.Getenv("SHELL"); s != "" {
		base := strings.ToLower(filepath.Base(s))
		switch base {
		case "zsh", "bash", "sh", "dash", "ksh", "fish", "pwsh", "powershell":
			return base
		}
	}
	return "zsh"
}

// shellRCFile 返回默认 shell 配置文件路径；--file 优先。
func shellRCFile(flag, shellName string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	switch shellName {
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	default:
		return filepath.Join(home, ".zshrc"), nil
	}
}

// envHookBlock 生成带标记的受管 hook 文本。
func envHookBlock(shellName string, shell proxyenv.Shell) string {
	var b strings.Builder
	fmt.Fprintln(&b, envHookStart)
	switch shell {
	case proxyenv.ShellFish:
		fmt.Fprintln(&b, "if command -v proxyctl >/dev/null 2>&1")
		fmt.Fprintf(&b, "  proxyctl env --shell %s 2>/dev/null | source\n", shellName)
		fmt.Fprintln(&b, "end")
	case proxyenv.ShellPowerShell:
		fmt.Fprintln(&b, "if (Get-Command proxyctl -ErrorAction SilentlyContinue) {")
		fmt.Fprintf(&b, "  Invoke-Expression (proxyctl env --shell %s 2>$null | Out-String)\n", shellName)
		fmt.Fprintln(&b, "}")
	default:
		fmt.Fprintln(&b, "if command -v proxyctl >/dev/null 2>&1; then")
		fmt.Fprintf(&b, "  eval \"$(proxyctl env --shell %s 2>/dev/null)\"\n", shellName)
		fmt.Fprintln(&b, "fi")
	}
	fmt.Fprintln(&b, envHookEnd)
	return b.String()
}

// installEnvHook 把 hook 写入配置文件；已存在时返回 (false, nil)。
func installEnvHook(rcPath, shellName string, shell proxyenv.Shell) (bool, error) {
	text, err := readFileOrEmpty(rcPath)
	if err != nil {
		return false, err
	}
	if strings.Contains(text, envHookStart) {
		return false, nil
	}
	if err := backupRCFile(rcPath); err != nil {
		return false, err
	}
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if err := os.WriteFile(rcPath, []byte(text+envHookBlock(shellName, shell)), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// removeEnvHook 移除受管 hook；未找到时返回 (false, nil)。
func removeEnvHook(rcPath string) (bool, error) {
	text, err := readFileOrEmpty(rcPath)
	if err != nil {
		return false, err
	}
	updated, ok := removeManagedBlock(text)
	if !ok {
		return false, nil
	}
	if err := os.WriteFile(rcPath, []byte(updated), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// removeManagedBlock 删除两个标记之间的内容。
func removeManagedBlock(text string) (string, bool) {
	start := strings.Index(text, envHookStart)
	if start < 0 {
		return text, false
	}
	end := strings.Index(text[start:], envHookEnd)
	if end < 0 {
		end = len(text) - start
	} else {
		end = start + end + len(envHookEnd)
	}
	updated := text[:start] + strings.TrimPrefix(text[end:], "\n")
	updated = strings.TrimRight(updated, "\n") + "\n"
	if updated == "\n" {
		updated = ""
	}
	return updated, true
}

// backupRCFile 首次安装前把原文件备份为 <文件>.proxyctl.bak（已存在则不覆盖）。
func backupRCFile(rcPath string) error {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	bak := rcPath + ".proxyctl.bak"
	if _, err := os.Stat(bak); err == nil {
		return nil
	}
	return os.WriteFile(bak, data, 0o644)
}

func readFileOrEmpty(p string) (string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func init() {
	envCmd.Flags().StringVar(&envShellFlag, "shell", "", "目标 shell（zsh/bash/sh/fish/powershell，默认按 $SHELL 推断）")
	envCmd.Flags().BoolVar(&envClearFlag, "clear", false, "输出直连模式（unset 全部代理变量）")
	envInstallCmd.Flags().StringVar(&envFileFlag, "file", "", "目标 shell 配置文件（默认 ~/.zshrc 等）")
	envRemoveCmd.Flags().StringVar(&envFileFlag, "file", "", "目标 shell 配置文件（默认 ~/.zshrc 等）")
	envCmd.AddCommand(envInstallCmd, envRemoveCmd)
}
