# proxyctl

proxyctl 是一个系统代理与端口管理命令行工具：查看系统代理状态、一键同步 git 全局代理、执行网络连通性测试、检查端口占用并结束占用进程。

变更历史见 [CHANGELOG.md](CHANGELOG.md)。

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
  -X github.com/alexwoo79/go_coding/proxyctl/cmd.version=1.0.0 \
  -X github.com/alexwoo79/go_coding/proxyctl/cmd.commit=$(git rev-parse --short HEAD) \
  -X github.com/alexwoo79/go_coding/proxyctl/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o proxyctl .
```

未注入版本信息时，`version` 命令会显示 `dev` / `unknown`。

也可以直接安装到 `$GOBIN`：

```bash
go install github.com/alexwoo79/go_coding/proxyctl@latest
```

## 快速开始

```bash
proxyctl status              # 查看系统代理与 git 代理状态
proxyctl apply               # 按系统代理自动设置 git 全局代理
proxyctl clear               # 关闭系统代理并清除 git 全局代理
proxyctl restore             # 从快照恢复 apply 前的系统代理与 git 代理
proxyctl test                # 网络连通性测试
proxyctl doctor              # 一键诊断开发环境网络状态
proxyctl env                 # 生成/安装终端代理环境变量
proxyctl tools               # 管理 npm/cargo/pip/docker 的代理配置
proxyctl status --json       # 以 JSON 输出状态（便于 AI Agent 消费）
proxyctl profile list        # 列出系统代理 profile
proxyctl port 7892           # 查看 7892 端口占用
proxyctl port 7892 --kill    # 结束占用 7892 端口的进程
```

## 命令说明

| 命令 | 说明 |
| --- | --- |
| `status` | 查看系统代理（HTTP/HTTPS/SOCKS/PAC）与 git 全局代理（`http.proxy`/`https.proxy`）状态 |
| `apply` | 读取系统代理并设置 git 全局代理；优先使用 HTTP 代理，未启用时回退到 SOCKS 代理 |
| `clear` | 关闭系统代理（HTTP/HTTPS/SOCKS/PAC）并删除 git 全局代理配置 |
| `restore` | 从 `apply` 保存的状态快照恢复系统代理与 git 全局代理 |
| `test` | 依次执行公网 IP、HTTP 响应、ping、git 连通性测试 |
| `doctor` | 一键诊断：系统代理、git 代理、端口监听、连通性、环境变量，发现问题时退出码为 1 |
| `env` | 把当前系统代理生成为终端环境变量脚本，或安装到 shell 配置文件 |
| `tools` | 读写 npm/cargo/pip/docker 各自的代理配置文件（apply/clear/restore/list） |
| `port` | 检查 TCP 端口占用，可结束占用进程 |
| `profile` | 保存 / 列出 / 应用 / 删除系统代理 profile |
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
| `--yes` | 配合 `--kill` 跳过结束确认（非交互式环境必须使用） |
| `--force` | 配合 `--kill` 使用，强制结束进程（Unix 发送 `kill -9`，Windows 使用 `taskkill /F`） |

端口号必须是 1–65535 之间的数字。

`--kill` 默认会在终端中要求确认；在脚本等非交互式环境中必须显式追加 `--yes`。

### 状态快照与恢复

执行 `proxyctl apply` 前，proxyctl 会把当前系统代理（每个网络服务的
HTTP/HTTPS/SOCKS/PAC 状态）与 git 全局代理保存到
`~/.config/proxyctl/state.json`（可用环境变量 `PROXYCTL_STATE_FILE` 覆盖）。

```bash
proxyctl clear     # 关闭系统代理并清除 git 代理（快照保留）
proxyctl restore   # 把系统代理与 git 代理恢复为 apply 前的状态
```

这样 `clear` 不会永久丢失你原有的代理配置。

### status --json

`status --json` 输出结构化 JSON，便于脚本或 AI Agent 消费：

```json
{
  "system_proxy": {
    "http": { "enabled": true, "host": "127.0.0.1", "port": "7890" },
    "pac": { "enabled": true, "url": "http://127.0.0.1:8080/proxy.pac" }
  },
  "git": {
    "http_proxy": "http://127.0.0.1:7890",
    "https_proxy": "http://127.0.0.1:7890"
  }
}
```

未启用的代理项不会出现在输出中；未设置的 git 配置项为 `null`。

### doctor

`doctor` 把 `status` / `test` / `port` 组合成一次完整诊断：

```console
$ proxyctl doctor
System / Proxy / Git / Ports / Connectivity / Environment ...
Recommendation
  → 端口 7890 未监听，但 git http.proxy 指向它：请确认代理程序已启动，或运行 proxyctl clear
```

存在问题时退出码为 1，适合接入脚本与 Agent 工作流。

### profile

profile 保存的是系统代理端点配置（HTTP/HTTPS/SOCKS/PAC），存放在
`~/.config/proxyctl/profiles/`（可用环境变量 `PROXYCTL_PROFILE_DIR` 覆盖）。

```bash
proxyctl profile save clash   # 保存当前系统代理为 profile "clash"
proxyctl profile list         # 列出全部 profile
proxyctl profile use clash    # 把 profile 应用到系统代理
proxyctl profile use direct   # 关闭系统代理（直连）
proxyctl profile remove clash # 删除 profile
```

`use` 只修改系统代理；如需同步 git 代理，随后执行 `proxyctl apply`。

### env（终端环境变量代理）

macOS 系统代理只对 GUI 程序生效；终端 CLI（curl、pip、npm、cargo、uv、
brew 等）遵循环境变量 `http_proxy`/`https_proxy`/`all_proxy`/`no_proxy`。
`env` 把两者打通：

```bash
eval "$(proxyctl env)"        # 当前终端立即走代理
eval "$(proxyctl env --clear)" # 当前终端立即直连
proxyctl env install          # 写入 ~/.zshrc（可用 --file 指定），新开终端自动生效
proxyctl env remove           # 移除 install 写入的 hook
```

`install` 写入的是受管 hook，每次新开终端自动执行 `proxyctl env`，
因此会跟随当前系统代理状态。配合 profile 即可实现不同网络环境一键切换：

```bash
proxyctl profile use clash   # 切到 Clash 环境（新开终端自动走代理）
proxyctl profile use direct  # 切到直连（新开终端自动恢复直连）
```

支持 `--shell zsh|bash|sh|fish|powershell`（默认按 `$SHELL` 推断）。
首次 `install` 前会把原配置文件备份为 `<文件>.proxyctl.bak`。

### tools（开发工具代理配置）

`env` 覆盖环境变量型工具；npm、pip、cargo、docker 还有各自的持久化配置文件，
由 `tools` 管理：

```bash
proxyctl tools list                # 查看各工具当前代理配置
proxyctl tools apply               # 把当前系统代理写入全部工具（可选指定工具）
proxyctl tools clear               # 清除工具代理配置
proxyctl tools restore             # 从快照恢复（apply/clear 前自动保存快照）
proxyctl tools apply npm cargo     # 只写指定工具
```

对应配置文件：

| 工具 | 配置文件 | 配置项 |
| --- | --- | --- |
| npm | `~/.npmrc`（可用 `npm_config_userconfig` 覆盖） | `proxy` / `https-proxy` |
| pip | `~/.config/pip/pip.conf`（可用 `PIP_CONFIG_FILE` 覆盖） | `[global] proxy` |
| cargo | `~/.cargo/config.toml`（可用 `CARGO_HOME` 覆盖） | `[http] proxy` |
| docker | `~/.docker/config.json`（可用 `DOCKER_CONFIG` 覆盖） | `proxies.default.httpProxy/httpsProxy/noProxy` |

`proxyctl clear` 会一并清除这些工具的代理配置；`proxyctl restore`
会一并恢复（前提是快照存在）。

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
