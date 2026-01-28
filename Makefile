# Makefile for regex-tester

# 项目名称
APP_NAME := regex-tester
VERSION := 1.0.1

# 构建目录
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin
DIST_DIR := $(BUILD_DIR)/dist

# Go 编译参数
GO := go
GO_BUILD_FLAGS := -ldflags="-s -w"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 默认目标
.PHONY: all
all: build

# 显示帮助信息
.PHONY: help
help:
	@echo "可用目标:"
	@echo "  build      - 编译程序 (默认目标)"
	@echo "  build-all  - 编译所有平台版本"
	@echo "  package    - 打包成 tar 包"
	@echo "  clean      - 清理构建文件"
	@echo "  test       - 运行测试"
	@echo "  help       - 显示此帮助信息"
	@echo ""
	@echo "环境变量:"
	@echo "  GOOS        - 目标操作系统 (默认: $(GOOS))"
	@echo "  GOARCH      - 目标架构 (默认: $(GOARCH))"
	@echo ""
	@echo "示例:"
	@echo "  make build              # 编译当前平台"
	@echo "  make GOOS=linux build   # 编译 Linux 版本"
	@echo "  make build-all          # 编译所有平台"
	@echo "  make package            # 打包发布"
	@echo "  make clean              # 清理构建文件"

# 创建构建目录
$(BUILD_DIR) $(BIN_DIR) $(DIST_DIR):
	@mkdir -p $@

# 编译当前平台
.PHONY: build
build: $(BIN_DIR)
	@echo "编译 $(APP_NAME) for $(GOOS)/$(GOARCH)..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/$(APP_NAME) main.go
	@echo "编译完成: $(BIN_DIR)/$(APP_NAME)"

# 编译所有支持的平台
.PHONY: build-all
build-all: $(BIN_DIR)
	@echo "编译所有平台版本..."
	@$(MAKE) build GOOS=linux GOARCH=amd64
	@$(MAKE) build GOOS=linux GOARCH=arm64
	@$(MAKE) build GOOS=darwin GOARCH=amd64
	@$(MAKE) build GOOS=darwin GOARCH=arm64
	@$(MAKE) build GOOS=windows GOARCH=amd64
	@echo "所有平台编译完成"

# 创建发布包
.PHONY: package
package: $(DIST_DIR) build
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)/$(APP_NAME)-$(VERSION)
	@cp $(BIN_DIR)/$(APP_NAME) $(DIST_DIR)/$(APP_NAME)-$(VERSION)/
	@cp config.json $(DIST_DIR)/$(APP_NAME)-$(VERSION)/
	@cp test_case.txt $(DIST_DIR)/$(APP_NAME)-$(VERSION)/
	@cp README.md $(DIST_DIR)/$(APP_NAME)-$(VERSION)/ 2>/dev/null || true
	@echo "版本: $(VERSION)" > $(DIST_DIR)/$(APP_NAME)-$(VERSION)/VERSION
	@echo "构建时间: $(shell date)" >> $(DIST_DIR)/$(APP_NAME)-$(VERSION)/VERSION
	@echo "平台: $(GOOS)/$(GOARCH)" >> $(DIST_DIR)/$(APP_NAME)-$(VERSION)/VERSION
	@cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz $(APP_NAME)-$(VERSION)
	@rm -rf $(DIST_DIR)/$(APP_NAME)-$(VERSION)
	@echo "发布包已创建: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz"

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@$(GO) test -v ./... || (echo "测试失败"; exit 1)
	@echo "测试通过"

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"