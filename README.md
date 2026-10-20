# modstart-cli

通过后台生成的凭证登录 ModStart 站点，直接调用站点后台已有接口的命令行工具。

- 所有命令统一输出 JSON（中文不转义），适合脚本与 AI 工具解析
- 支持多站点、多后台路径
- CLI 接口统一在 `/{adminPath}/cli/*`，由各模块 `AdminCli`（控制器 + routes + api.json）提供

## 给 AI 的安装提示词

把下面这段发给 AI（替换 `<后台管理地址>` 与 `<凭证>`）：

```text
请帮我安装并使用 modstart-cli：

TODO

注意：
- `--admin-url` 要填完整后台管理地址（含 /admin 等后台路径），不要只填域名。
- 不要猜测接口 key，先用 `api list` 查询。
- 不要修改站点后台代码。
```

## 安装

### 直接下载（推荐）

推送 `main` 后由 GitHub Actions 自动打包，产物发布在 `latest`：

```text
https://github.com/open-tecmz/modstart-cli/releases/latest
```

| 平台 | 压缩包 |
| --- | --- |
| macOS（Apple Silicon） | `modstart-cli-darwin-arm64.tar.gz` |
| macOS（Intel） | `modstart-cli-darwin-amd64.tar.gz` |
| Linux x86_64 | `modstart-cli-linux-amd64.tar.gz` |
| Linux arm64 | `modstart-cli-linux-arm64.tar.gz` |
| Windows x86_64 | `modstart-cli-windows-amd64.zip` |
| Windows arm64 | `modstart-cli-windows-arm64.zip` |

```shell
# macOS Apple Silicon 示例，其它平台替换 PLATFORM 即可
OWNER=open-tecmz
PLATFORM=darwin-arm64
curl -L -o modstart-cli.tar.gz "https://github.com/${OWNER}/modstart-cli/releases/latest/download/modstart-cli-${PLATFORM}.tar.gz"
tar xzf modstart-cli.tar.gz
sudo mv modstart-cli /usr/local/bin/
modstart-cli version
```

- 压缩包内即单个可执行文件（解压后直接是 `modstart-cli` / `modstart-cli.exe`）
- `latest` 里同时提供未压缩的可执行文件与 `checksums.txt`（sha256 校验）
- 滚动版本号形如 `0.1.0+<提交短 SHA>`；正式版本以 `vX.Y.Z` 标签发布（同样标记为 Latest）

### 从源码编译

```shell
make update-version 0.1.0   # 设置版本号（也可用 make update-version VERSION=0.1.0）
make dev    # 本机编译，产物 build/modstart-cli
make        # 多平台编译（darwin/linux/windows × amd64/arm64）
make dist   # 多平台编译 + 打包 dist/（压缩包、裸可执行文件、checksums.txt）
```

> `make update-version` 不带参数时生成时间版本号（`vYYYYMMDDHHMMSS`）；
> `make` / `make dev` 不会改动版本号。

## 快速开始

### 1. 后台生成凭证

后台进入「系统设置 → 接口设置 → CLI 授权管理」，点击「新增」：

- 填写设备名称
- **授权接口**树状勾选（模块 → 具体接口，必须显式勾选，仅创始人可创建凭证）
- 有效期默认 90 天

提交后自动弹出凭证复制窗口（仅显示一次）。

### 2. 登录

```shell
modstart-cli login --admin-url=https://demo.com/admin --token=mscli_xxxxxxxx --name=MacBook
```

> `--admin-url` 填**完整后台管理地址**。后台路径可能被修改（如 `/manage`），
> CLI 会自动拆分为「站点根地址 + 后台路径」，后续所有请求都基于该地址。

### 3. 使用

```shell
modstart-cli whoami                    # 当前登录管理员
modstart-cli api list                  # 可用接口（JSON）
modstart-cli call member.index '{"page":1,"pageSize":20}'
modstart-cli call news.index '{"page":1,"pageSize":20}'
modstart-cli call news.add '{"title":"标题","content":"内容"}'
modstart-cli call news.category.index
modstart-cli call news.edit '{"id":1,"title":"新标题"}'
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `login --admin-url=... --token=... [--name=...]` | 校验凭证并保存（服务端无登录接口，无 Cookie） |
| `logout` | 清除本地凭证 |
| `whoami` | 当前登录管理员 |
| `site list` / `site use <后台地址>` / `site remove <后台地址>` | 多站点管理 |
| `api list` | 列出可用接口与参数 |
| `call <key> ['<json>']` | 调用接口，参数为 JSON 对象（也可用 `--data '<json>'`），会按接口定义在本地校验 |
| `version` | 版本 |

全局参数：

- `--site <站点根地址>`：临时指定站点，默认使用当前站点

## 输出格式

- 所有命令统一输出 JSON，中文不转义（不会出现 `\u64cd\u4f5c`）
- `call` 固定返回 `{ code, msg, data }`（登录跳转时含 `redirect`）
- `api list` 返回 `{ adminPrefix, apis }`
- `site list` 返回 `{ current, sites }`（不含 token）
- 可配合 `jq` 使用，例如 `modstart-cli api list | jq '.apis[].key'`

## 多站点

配置保存在 `~/.modstart-cli/config.json`（目录 0700、文件 0600），按站点根地址分区：

```json
{
  "current": "https://demo.com",
  "sites": {
    "https://demo.com": {
      "url": "https://demo.com",
      "adminUrl": "https://demo.com/admin",
      "adminPath": "admin",
      "name": "MacBook",
      "token": "mscli_xxx"
    }
  }
}
```

> `site list` 的 JSON 输出不会包含 token，可安全分享。


## 常见问题

| 现象 | 处理 |
| --- | --- |
| `接口不存在（HTTP 404）` | 确认 `--admin-url` 是完整后台地址、站点已启用 AdminCli 模块；必要时清理站点缓存 |
| 返回凭证失效 | CLI 用本地凭证自动重试；凭证被撤销时需在后台重新生成 |
| 看不到某个接口 / 调用被拒绝 | 确认对应模块存在 `AdminCli/api.json`，且凭证创建时已勾选该接口（未勾选时请求会被直接拒绝） |
| 中文显示为 `\uXXXX` | 升级到最新版本（JSON 输出已关闭 Unicode 转义） |

## 开发

```shell
make update-version 0.1.0   # 更新版本号（不传参则使用时间版本号）
make dev    # 本机编译
make dist   # 多平台编译 + 打包 dist/
make test   # 测试
go build ./...
```

## 发布

`.github/workflows/release.yml`：

| 触发 | 结果 |
| --- | --- |
| 推送 `main` | 重新构建并覆盖 `latest` 发布（标记为 Latest），版本号 = `0.1.0+<提交短 SHA>` |
| 推送 `v*.*.*` 标签 | 发布同名版本（标记为 Latest） |
| 手动运行 workflow | 同 `main` |

每次发布包含 6 个平台的压缩包与裸可执行文件，以及 `checksums.txt`：

```
modstart-cli-darwin-amd64(.tar.gz)   modstart-cli-linux-amd64(.tar.gz)   modstart-cli-windows-amd64(.exe|.zip)
modstart-cli-darwin-arm64(.tar.gz)   modstart-cli-linux-arm64(.tar.gz)   modstart-cli-windows-arm64(.exe|.zip)
checksums.txt
```

本地复现同样流程：`make update-version 0.1.0 && make dist`，产物在 `dist/`。
