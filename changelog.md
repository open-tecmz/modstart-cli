# changelog


## 0.1.0

- 新增 GitHub Actions 发布流程（`.github/workflows/release.yml`）：推送 `main` 自动构建各平台二进制并滚动更新 `latest` 发布，推送 `v*.*.*` 标签发布正式版本
- 发布产物包含 darwin/linux/windows × amd64/arm64 的压缩包、裸可执行文件与 `checksums.txt`（sha256），压缩包内附 `LICENSE` 与 `README.md`
- 新增 `make dist`：本地交叉编译并打包到 `dist/`
- 多平台编译的 Windows 产物改为带 `.exe` 后缀
- 项目采用 Apache License 2.0 开源（新增 `LICENSE`）
- 新增 `make update-version [version]`：指定版本号写入 `lib/version/version.go`，不传参数时生成时间版本号
- 版本号统一为 `0.1.0`
- `make` / `make dev` 不再自动改写版本号，版本只在显式执行 `update-version` 时变化

## 1.0.0

- 第一个版本
