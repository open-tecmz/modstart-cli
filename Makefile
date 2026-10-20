# 编译目标
TARGET = modstart-cli

# 源文件目录
SRCDIR = .

# 编译输出目录
BINDIR = build

# 发布包目录
DISTDIR = dist

# 发布平台（GOOS/GOARCH）
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

# 构建命令
GO = go
BUILD = CGO_ENABLED=0 $(GO) build

# 源文件
SOURCES := $(shell find $(SRCDIR) -name '*.go')

BUILD_TIME := $(shell date "+%Y-%m-%d %H:%M:%S")
BUILD_VERSION := $(shell date "+v%Y%m%d%H%M%S")

# 版本号来源：
#   make update-version              使用时间版本号（BUILD_VERSION）
#   make update-version 0.1.0        使用位置参数
#   make update-version VERSION=0.1.0 使用变量（命令行变量优先级最高）
VERSION := $(strip $(filter-out update-version,$(MAKECMDGOALS)))

# 位置参数（如 0.1.0）只表示版本号，需要一个空的兜底规则避免 make 报错
ifneq ($(VERSION),)
%:
	@:
endif

.PHONY: default update-version build dist dev test vet clean

# 默认目标：清理 → 多平台编译（不再改写版本号，版本由 update-version 维护）
default: clean build

# 更新 lib/version/version.go 中的版本号与构建时间
update-version:
	@v="$(VERSION)"; \
	if [ -z "$$v" ]; then v="$(BUILD_VERSION)"; fi; \
	printf '%s' "$$v" | grep -Eq '^[0-9A-Za-z][0-9A-Za-z._+-]*$$' \
		|| { echo "版本号格式不正确：$${v}（示例：0.1.0、v1.2.3）" >&2; exit 1; }; \
	printf 'package version\n\n// 构建时由 Makefile 自动写入\nvar VERSION = "%s"\nvar BUILD_TIME = "%s"\n' \
		"$$v" "$(BUILD_TIME)" > lib/version/version.go; \
	echo "版本已更新：$${v}（$(BUILD_TIME)）"

# 多平台编译（Windows 产物带 .exe 后缀）
build:
	@mkdir -p $(BINDIR)
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; \
		arch=$${p#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "  build $(BINDIR)/$(TARGET)-$$os-$$arch$$ext"; \
		GOOS=$$os GOARCH=$$arch $(BUILD) -trimpath -o $(BINDIR)/$(TARGET)-$$os-$$arch$$ext $(SRCDIR) || exit 1; \
	done

# 打包各平台压缩包（tar.gz / zip）+ 裸可执行文件 + checksums.txt，产物在 dist/
dist: build
	@./scripts/package.sh $(BINDIR) $(DISTDIR) $(TARGET)

# 本机编译（开发调试）
dev:
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 $(GO) build -trimpath -o $(BINDIR)/$(TARGET) $(SRCDIR)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

# 清理
clean:
	@rm -rf $(BINDIR)
