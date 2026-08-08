# proxyctl

proxyctl 是一个系统代理与端口管理命令行工具：查看系统代理状态、一键同步 git 全局代理、执行网络连通性测试、检查端口占用并结束占用进程。

```console
$ proxyctl status
=== macOS 系统代理 ===
HTTP   代理: 未启用
HTTPS  代理: 未启用
SOCKS  代理: 未启用
PAC    自动代理: 未启用

=== Git 全局代理 ===
  http.proxy  = http://127.0.0.1:7892
  https.proxy = http://127.0.0.1:7892
```

## 安装

### 从源码构建

需要 Go 1.25 或更高版本：

```bash
go build -o proxyctl .
```

建议在构建时注入版本信息（版本号、commit、构建时间）：

```bash
go build -ldflags "\
  -X github.com/alexwoo79/go_coding/cobra/cmd.version=1.0.0 \
  -X github.com/alexwoo79/go_coding/cobra/cmd.commit=$(git rev-parse --short HEAD) \
  -X github.com/alexwoo79/go_coding/cobra/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o proxyctl .
```

未注入版本信息时，`version` 命令会显示 `dev` / `unknown`。

也可以直接安装到 `$GOBIN`：

```bash
go install github.com/alexwoo79/go_coding/cobra@latest
```

## 快速开始

```bash
proxyctl status              # 查看系统代理与 git 代理状态
proxyctl apply               # 按系统代理自动设置 git 全局代理
proxyctl clear               # 关闭系统代理并清除 git 全局代理
proxyctl test                # 网络连通性测试
proxyctl port 7892           # 查看 7892 端口占用
proxyctl port 7892 --kill    # 结束占用 7892 端口的进程
```

## 命令说明

| 命令 | 说明 |
| --- | --- |
| `status` | 查看系统代理（HTTP/HTTPS/SOCKS/PAC）与 git 全局代理（`http.proxy`/`https.proxy`）状态 |
| `apply` | 读取系统代理并设置 git 全局代理；优先使用 HTTP 代理，未启用时回退到 SOCKS 代理 |
| `clear` | 关闭系统代理（HTTP/HTTPS/SOCKS/PAC）并删除 git 全局代理配置 |
| `test` | 依次执行公网 IP、HTTP 响应、ping、git 连通性测试 |
| `port` | 检查 TCP 端口占用，可结束占用进程 |
| `version` | 显示版本、commit、构建时间与 Go 版本 |
| `completion` | 生成 bash / zsh / fish / powershell 自动补全脚本 |

运行 `proxyctl help <命令>` 可查看每个命令的详细说明。

### port

```bash
proxyctl port <端口号> [--all] [--kill] [--force]
```

| 标志 | 说明 |
| --- | --- |
| `--all` | 显示该端口的全部连接（默认仅显示 LISTEN 监听进程） |
| `--kill` | 结束占用该端口的进程 |
| `--force` | 配合 `--kill` 使用，强制结束进程（Unix 发送 `kill -9`，Windows 使用 `taskkill /F`） |

端口号必须是 1–65535 之间的数字。

### apply 与当前终端

`apply` 设置的是 git 全局配置，只对之后的 git 命令生效。子进程无法修改父 shell 的环境变量，因此若要让当前终端里的 HTTP 请求走代理，需手动执行命令输出的 `export` 提示。

## 平台支持

| 功能 | macOS | Windows | 其他 Unix |
| --- | --- | --- | --- |
| `status` / `apply` / `clear` | ✓ | ✓（注册表） | ✗ |
| `test` | ✓ | ✓ | ✓（无系统代理检测，回退环境变量/直连） |
| `port` | ✓ | ✓ | ✓（需 `lsof`） |

依赖的外部命令：

- macOS：`scutil`、`networksetup`、`lsof`（系统自带）
- Windows：`netstat`、`tasklist`、`taskkill`、`powershell`（系统自带）
- 其他 Unix：`lsof`（macOS 自带；Linux 可用 `apt install lsof` 或 `yum install lsof` 安装）
- 所有平台：`git`、`ping`

`test` 命令的 HTTP 请求优先使用检测到的系统代理（HTTP → HTTPS → SOCKS），未检测到时回退到环境变量代理或直连。

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 成功 |
| `1` | 运行时错误（如无法读取系统代理、进程结束失败） |
| `2` | 用法错误（无效标志、无效参数、未知子命令） |

## 开发

```bash
go test ./...    # 运行单元测试
go vet ./...     # 静态检查
gofmt -l .       # 检查代码格式（应无输出）
```

## License

项目目前未指定开源许可证，仓库中暂无 `LICENSE` 文件。
