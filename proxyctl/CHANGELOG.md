# Changelog

本项目按里程碑记录变更。当前处于开发阶段，版本为 `dev`，所有变更均未发布、
未打 tag、未提交到远端。

## [dev] 2026-08-11 — 架构重构与功能完善

基于一轮代码级 review 的结论，对 proxyctl 做了 P0 安全修复与功能扩展，
目标是把它定位为「开发环境网络控制 CLI」：终端命令可自由开启代理或直连，
并适配不同网络环境。

### 新增命令

| 命令 | 说明 |
| --- | --- |
| `restore` | 从快照恢复 `apply` 前的系统代理与 git 全局代理 |
| `doctor` | 一键诊断：系统代理、git 代理、端口监听、连通性、环境变量，发现问题退出码 1 |
| `env` | 把系统代理生成为终端环境变量脚本（支持 zsh/bash/sh/fish/powershell） |
| `env install` / `env remove` | 安装 / 移除 shell 受管 hook，新开终端自动跟随系统代理 |
| `tools list/apply/clear/restore` | 读写 npm / pip / cargo / docker 各自的代理配置文件 |
| `profile list/save/use/remove` | 系统代理 profile 管理（内置 `direct` 直连） |
| `status --json` | 结构化 JSON 输出，便于脚本与 AI Agent 消费 |

### 安全与正确性修复

- **快照 / 恢复**：`apply` 前保存系统代理（逐网络服务）与 git 全局代理到
  `~/.config/proxyctl/state.json`；`clear` 不再永久丢失原配置，`restore` 可完整恢复。
- **`--kill` 确认**：`port --kill` 默认交互确认；非交互环境必须显式 `--kill --yes`。
- **git 错误不再被吞掉**：新增 `internal/git`，区分「未设置」与「读取失败」，
  不再把 git 不存在 / 配置损坏伪装成空字符串。
- **PAC URL 保留**：`Info` 新增 `AutoConfigURL`，`status` 现在能显示 PAC 脚本地址。

### 架构调整

- 新增包：`internal/git`、`internal/state`、`internal/diagnostic`、
  `internal/profile`、`internal/proxyenv`、`internal/toolproxy`。
- `cmd` 层瘦身：git 操作、网络诊断、profile 存储、环境变量生成、工具配置读写
  全部下沉到 internal 包，`cmd` 只做命令编排与输出。
- `test` 重构为诊断框架：`Checker` 接口 + `PublicIPCheck` / `HTTPCheck` /
  `PingCheck` / `GitCheck`，每步带耗时。
- port 模型修正：`Process.PID` 改为 `int`（数字排序），一个进程的多个监听地址
  合并展示；lsof / netstat 解析器独立并补充测试。
- 系统代理快照 / 恢复与 profile 应用共用同一套 `applyToService` 逻辑（macOS），
  Windows 走注册表。

### 终端代理打通

- `env`：生成 `http_proxy` / `https_proxy` / `all_proxy` / `no_proxy`
  （含大写形式）脚本，覆盖 curl、pip、npm、cargo、uv、brew 等 CLI；
  `--clear` 输出直连（unset 全部）。
- `env install`：写入 `~/.zshrc`（可用 `--file` 指定）的受管 hook，
  新开终端自动执行 `proxyctl env`，配合 `profile use` 实现不同网络环境一键切换；
  首次安装自动备份为 `<文件>.proxyctl.bak`，重复安装幂等。
- `tools`：npm（`~/.npmrc`）、pip（`pip.conf [global] proxy`）、
  cargo（`config.toml [http] proxy`）、docker（`config.json proxies.default.*`），
  写入前保留原有无关配置、原子写入、快照可恢复；`clear` / `restore` 主命令
  也会一并处理这些工具。

### 测试

- 新增解析器测试：`scutil`、`networksetup`、lsof、INI / TOML / JSON 工具配置。
- 新增单元测试：git 假 runner、state / profile / tools 快照存取、
  env 脚本与 hook 安装 / 移除 / 幂等 / 备份、JSON builder、诊断编排。
- 工程化：`gofmt` / `go vet` / `go test ./...` 全绿；
  Windows / Linux 交叉编译通过。

### 配置文件与路径

| 用途 | 默认路径 | 可覆盖环境变量 |
| --- | --- | --- |
| apply 快照 | `~/.config/proxyctl/state.json` | `PROXYCTL_STATE_FILE` |
| tools 快照 | `~/.config/proxyctl/tools-state.json` | `PROXYCTL_TOOLS_STATE_FILE` |
| profiles | `~/.config/proxyctl/profiles/` | `PROXYCTL_PROFILE_DIR` |
| npm 配置 | `~/.npmrc` | `npm_config_userconfig` |
| pip 配置 | `~/.config/pip/pip.conf` | `PIP_CONFIG_FILE` |
| cargo 配置 | `~/.cargo/config.toml` | `CARGO_HOME` |
| docker 配置 | `~/.docker/config.json` | `DOCKER_CONFIG` |

### 当前状态与后续计划

- 所有改动位于工作区，尚未提交、未发布；版本仍为 `dev`。
- 进入稳定期：先观察使用并修复问题，暂不加新功能。
- 待办（稳定期后评估）：`test` / `doctor` / `tools list` 的 `--json` 输出、
  测试目标可配置（`--target` / 配置文件）、zsh 补全、正式版本号与 ldflags 注入。
