# AGENTS.md

# ModStart CLI（modstart-cli）

通过后台生成的凭证登录 ModStart 站点，直接调用站点后台已有接口的命令行工具。

## 基础规范

- 始终使用**简体中文**与用户交流。
- 指令不清晰时必须反问澄清，严禁幻觉。
- 极简原则：优先使用 Go 标准库，只有必要能力才引入第三方库。

## 项目结构

| 路径 | 说明 |
| --- | --- |
| `main.go` | 入口 |
| `command/` | cobra 命令（login/logout/whoami/site/api/call/version） |
| `lib/config/` | 本地配置与多站点凭证存储（`~/.modstart-cli/config.json`） |
| `lib/api/` | HTTP 客户端（登录、调用、Cookie、自动重登） |
| `lib/console/` | 终端输入输出 |
| `lib/version/` | 版本号（Makefile 自动写入） |
| `scripts/package.sh` | 打包各平台压缩包 + `checksums.txt` |
| `.github/workflows/release.yml` | 推送 `main` 自动发布 `latest`（各平台二进制） |
| `LICENSE` | Apache License 2.0 |
| `Makefile` | 多平台编译入口 |

## 核心机制

- 后台生成 CLI 凭证（`mscli_` 前缀），CLI 通过 `POST {AdminPath}/{app}_cli/login` 换取 Session Cookie；
- 登录使用 `--admin-url`（完整后台管理地址，如 `https://demo.com/admin`），自动拆分站点根地址与后台路径；
- 之后请求 `{AdminPath}/cli/<url>`（模块 CLI 控制器），由控制器委托后台原有 Admin Controller，不做映射转发；
- 模块 CLI 接口统一挂载在 `{AdminPath}/cli/*`，由各模块 `AdminCli`（控制器 + routes.php + api.json）提供；
- 接口 schema 由各模块 `AdminCli/api.json` 声明，CLI 通过 `{AdminPath}/cli/api_list` 读取；
- 身份维度用 `app` 表示，首期为 `admin`，`Site.App` 字段为未来 `member`/`web` 等预留。

## 输出约定

- **所有命令统一输出 JSON**，且中文不转义（`SetEscapeHTML(false)`，禁止出现 `\uXXXX`）；
- `call` 固定返回 `{ code, msg, data }`（登录跳转时含 `redirect`），不提供表格等多种格式；
- `call` 的输出包含 `code / msg / data / redirect`；
- `site list` 的 JSON 必须脱敏，禁止输出 token / cookie。

## 构建与发布

```shell
make update-version 0.1.0   # 更新版本号：make update-version [version]（不传则自动生成时间版本）
make        # 多平台编译(darwin/linux/windows × amd64/arm64)，Windows 产物带 .exe
make dist   # 多平台编译 + 打包 dist/（tar.gz / zip / 裸可执行文件 / checksums.txt）
make dev    # 仅编译本机版本，便于调试
make test   # 运行测试
```

- `make update-version [version]` 把版本号写入 `lib/version/version.go`：
  - `make update-version 0.1.0`（位置参数）或 `make update-version VERSION=0.1.0`（变量）
  - 不传版本号时自动生成时间版本（`vYYYYMMDDHHMMSS`）
  - 同时刷新 `BUILD_TIME`，版本号只允许字母、数字与 `. _ + -`
- `make` / `make dev` 不会改写版本号，版本只在显式执行 `update-version` 时变化。
- 编译产物输出到 `build/`，发布包输出到 `dist/`。
- 发布：推送 `main` 触发 `.github/workflows/release.yml`，用 `make dist` 交叉编译并滚动更新 `latest` 发布
  （版本号 = 源版本号 + `+<提交短 SHA>`）；推送 `vX.Y.Z` 标签则发布同名正式版本。
  产物命名 `modstart-cli-<os>-<arch>[.exe|.tar.gz|.zip]`，另有 `checksums.txt`（sha256），
  压缩包内附 `LICENSE` 与 `README.md`。
- 新增平台时同步修改 `Makefile` 的 `PLATFORMS` 与 `scripts/package.sh` 的 `PLATFORMS`。

## 开发规范

- 每次代码变更后必须执行 `go build ./...` 验证编译。
- 命名：导出符号 `PascalCase`，内部 `camelCase`，包/文件名全小写。
- 单文件建议不超过 500 行，超限必须拆分。
- 新增/修改命令后，必须同步更新 `README.md`、`skill/SKILL.md` 与 `changelog.md`。

## changelog.md 维护

- 每次代码变更必须同步更新根目录 `changelog.md`，追加 `- ` 列表项。
- 变更项聚焦用户可感知的能力/修复。
