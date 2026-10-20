---
name: modstart-cli
description: 使用 modstart-cli 通过后台生成的凭证登录 ModStart 站点，查询并调用站点后台已有接口（统一 JSON 输出）。当用户需要“用命令行/AI 管理 ModStart 后台”、查询某模块可用接口、调用后台 CRUD 或系统操作时使用。
---

# Skill: modstart-cli

## 适用场景

- 用户希望用命令行或 AI 管理 ModStart 站点后台（查看/新增/编辑/删除数据、清缓存、查管理员等）
- 需要先发现站点有哪些后台接口可调用
- 需要以脚本方式批量调用后台接口

## 前置条件

1. 站点已安装并启用 `AdminCli` 模块
2. 已在后台「系统设置 → 接口设置 → CLI 授权管理」生成凭证（`mscli_` 开头，仅显示一次）
3. 已知**完整后台管理地址**（如 `https://demo.com/admin`，后台路径可能被修改）

## 安装 CLI

优先下载预编译二进制（推送 `main` 后自动发布到 `latest`）：

```text
https://github.com/open-tecmz/modstart-cli/releases/latest
```

| 平台 | 压缩包 |
| --- | --- |
| macOS Apple Silicon | `modstart-cli-darwin-arm64.tar.gz` |
| macOS Intel | `modstart-cli-darwin-amd64.tar.gz` |
| Linux x86_64 / arm64 | `modstart-cli-linux-amd64.tar.gz` / `modstart-cli-linux-arm64.tar.gz` |
| Windows x86_64 / arm64 | `modstart-cli-windows-amd64.zip` / `modstart-cli-windows-arm64.zip` |

```shell
# macOS Apple Silicon 示例
OWNER=open-tecmz
curl -L -o modstart-cli.tar.gz "https://github.com/${OWNER}/modstart-cli/releases/latest/download/modstart-cli-darwin-arm64.tar.gz"
tar xzf modstart-cli.tar.gz
sudo mv modstart-cli /usr/local/bin/
modstart-cli version
```

- 压缩包内是可执行文件（`modstart-cli`，Windows 为 `modstart-cli.exe`）以及 `LICENSE`、`README.md`
- `latest` 里还有裸可执行文件与 `checksums.txt`（sha256）；滚动版本号形如 `0.1.0+<提交短 SHA>`
- 正式版本以 `vX.Y.Z` 标签发布

从源码编译（本机调试）：

```shell
cd modstart-cli
make dev        # 产物：build/modstart-cli
make dist       # 全部平台压缩包，产物在 dist/（含 checksums.txt）
```

> 下文示例中的 `./build/modstart-cli` 可替换为已安装到 PATH 的 `modstart-cli`。

## 登录

```shell
./build/modstart-cli login --admin-url=https://demo.com/admin --token=<凭证> --name=<设备名>
```

- 必须使用 `--admin-url`（完整后台地址），CLI 会自动拆分站点根地址与后台路径
- `login` 只是调用 `cli/me` 校验凭证并保存，服务端没有 login / logout 接口，也不使用 Cookie / Session
- 凭证保存在 `~/.modstart-cli/config.json`（0600），后续请求只带 `admin-cli-key`
- 凭证失效时 CLI 会自动用本地凭证重试

## 发现接口

```shell
./build/modstart-cli api list
```

返回 JSON：

```json
{
  "adminPrefix": "admin",
  "apis": [
    {
      "key": "news.index",
      "title": "新闻列表",
      "method": "post",
      "url": "news/index",
      "scope": "news.read",
      "params": [{ "name": "page", "type": "int", "required": false, "desc": "页码" }]
    }
  ]
}
```

**不要猜测 key**：必须先 `api list` 查询，再按 `params` 组装参数。

## 调用接口

```shell
./build/modstart-cli call news.index '{"page":1,"pageSize":20}'
./build/modstart-cli call news.add '{"title":"标题","content":"内容"}'
./build/modstart-cli call news.edit '{"id":1,"title":"新标题"}'
./build/modstart-cli call news.delete '{"ids":[4,5]}'
```

- 第二个位置参数统一为 **JSON 对象**，也支持 `--data '<JSON>'`（两者等价）
- 数字、布尔保持 JSON 类型；数组/对象会原样以 JSON 字符串传给服务端
- 必填参数缺失时 CLI 会直接报错并提示参数名
- 所有命令统一输出 JSON（中文不转义），`call` 固定返回 `{ code, msg, data }`
- `call` 会按接口定义在**本地校验参数**（必填、int/bool/json 类型），未声明参数会提示

## 常用命令

```shell
./build/modstart-cli whoami
./build/modstart-cli site list
./build/modstart-cli site use https://demo.com
./build/modstart-cli logout
```

## 约束

- **不要**为 CLI 新增业务接口：接口由模块 `AdminCli`（控制器 + routes + api.json）提供，控制器实现真实业务逻辑
- **不要**猜测接口或参数，先 `api list`
- 写操作（新增/编辑/删除）前先确认用户意图；删除类操作需用户明确要求
- 不要把 `~/.modstart-cli/config.json` 中的 token 输出到对话或日志
- `site list` 的 JSON 已脱敏（不含 token），可安全展示

## 常见问题

| 现象 | 处理 |
| --- | --- |
| `接口不存在（HTTP 404）` | 检查 `--admin-url` 是否为完整后台地址、站点是否启用 AdminCli、必要时清理站点缓存 |
| 看不到某接口 | 对应模块缺少 `AdminCli/api.json`，或凭证创建时未勾选该接口 |
| 中文显示 `\uXXXX` | 使用最新版本 CLI（JSON 输出已关闭 Unicode 转义） |
| 凭证失效 | 凭证可能被撤销，需在后台重新生成 |
